package uiworks

import (
	"encoding/json"
	"fmt"
	"github.com/fpawel/goutils/procmq"
	"time"

	"github.com/fpawel/ankat/internal/ankat"
	"github.com/fpawel/ankat/internal/db/worklog"
	"github.com/jmoiron/sqlx"
)

type Runner struct {
	delphiApp *procmq.ProcessMQ
	chStop,
	chInterrupt, chClose chan struct{}
	chStartMainTask,
	chNotifyWork chan notifyWork
	chGetInterrupted         chan chan bool
	chEndWork                chan endWork
	chUnsubscribeInterrupted chan chan struct{}
	chDelay                  chan delayInfo
	chDelaySkipped           chan struct{}
	chWriteLog               chan worklog.WriteRecord
	chStartWork              chan runWork
	chCurrentRunTask         chan chan *Task
}

type delayInfo struct {
	Name       string
	DurationMS int64
	Enabled    bool
}

type notifyWork struct {
	Ordinal int
	Name    string
	Run     bool
}

type endWork struct {
	Name  string
	Error error
}

type runWork struct {
	work Work
	n    int
	end  func()
}

func NewRunner(processMQ *procmq.ProcessMQ) Runner {
	x := Runner{
		delphiApp:                processMQ,
		chEndWork:                make(chan endWork),
		chStop:                   make(chan struct{}),
		chClose:                  make(chan struct{}),
		chInterrupt:              make(chan struct{}),
		chNotifyWork:             make(chan notifyWork),
		chStartMainTask:          make(chan notifyWork),
		chStartWork:              make(chan runWork),
		chGetInterrupted:         make(chan chan bool),
		chUnsubscribeInterrupted: make(chan chan struct{}),
		chDelay:                  make(chan delayInfo),
		chDelaySkipped:           make(chan struct{}),
		chWriteLog:               make(chan worklog.WriteRecord),
		chCurrentRunTask:         make(chan chan *Task),
	}

	processMQ.Handle("CURRENT_WORK_START", func(bytes []byte) interface {}{
		var payload notifyWork
		mustUnmarshalJson(bytes, &payload)
		x.chStartMainTask <- payload
		return nil
	})

	processMQ.Handle("SKIP_DELAY", func([]byte) interface {}{
		x.chDelaySkipped <- struct{}{}
		return nil
	})

	return x
}

func (x Runner) Close() {
	x.chClose <- struct{}{}
}

func (x Runner) CurrentRunTask() *Task {
	ch := make(chan *Task)
	x.chCurrentRunTask <- ch
	return <-ch
}

func (x Runner) Run(dbWork, dbConfig *sqlx.DB,  mainTask *Task) {

	rootTask := mainTask
	started := false
	interrupted := false
	closed := false
	workLogWritten := false

	var currentRunTask *Task

	writeLog := func(m worklog.WriteRecord) {
		if !workLogWritten {
			worklog.AddRootWork(dbWork, rootTask.name)
			workLogWritten = true
		}
		m.Works = currentRunTask.ParentKeys()
		x.delphiApp.Send("CURRENT_WORK_MESSAGE", worklog.Write(dbWork, m))
	}

	for {
		select {

		case ch := <-x.chCurrentRunTask:
			ch <- currentRunTask

		case <-x.chClose:

			if started {
				closed = true
				interrupted = true
			} else {
				return
			}

		case <-x.chInterrupt:
			interrupted = true

		case m := <-x.chWriteLog:
			writeLog(m)

		case v := <-x.chDelay:
			x.delphiApp.Send("DELAY", v)

		case d := <-x.chStartWork:
			if started {
				panic("run twice")
			}
			rootTask = d.work.Task()
			task := rootTask.descendants[d.n]
			currentRunTask = task

			//x.delphiApp.Send("SETUP_CURRENT_WORKS", task.Info(dbLog))
			started = true
			interrupted = false
			workLogWritten = false
			go func() {
				x.notifyWork(task, true)
				err := task.perform(x, dbConfig)
				d.end()
				x.chEndWork <- endWork{
					Error: err,
					Name:  task.name,
				}
				x.notifyWork(task, false)
			}()

		case rm := <-x.chStartMainTask:
			if started {
				panic("run twice")
			}
			rootTask = mainTask
			m := rootTask.GetTaskByOrdinal(rm.Ordinal)
			if !m.Checked(dbWork) {
				x.delphiApp.Send("ERROR", struct {
					Text string
				}{m.Text() + ": операция не отмечена"})
				continue
			}

			started = true
			interrupted = false
			workLogWritten = false
			go func() {
				x.chEndWork <- endWork{
					Error: m.perform(x, dbConfig),
					Name:  m.name,
				}
			}()

		case ch := <-x.chGetInterrupted:
			ch <- interrupted

		case v := <-x.chEndWork:
			if closed {
				return
			}
			started = false
			interrupted = true
			currentRunTask = nil
			x.delphiApp.Send("END_WORK", struct {
				Name, Error string
			}{v.Name, fmtErr(v.Error)})

		case <-x.chStop:
			if started {
				interrupted = true
			}

		case r := <-x.chNotifyWork:
			x.delphiApp.Send("CURRENT_WORK", r)
			if r.Run {
				currentRunTask = rootTask.GetTaskByOrdinal(r.Ordinal)
			}
		}
	}
}

func (x Runner) notifyWork(m *Task, run bool) {
	x.chNotifyWork <- notifyWork{
		Ordinal: m.ordinal,
		Name:    m.name,
		Run:     run,
	}
}

func (x Runner) Perform(ordinal int, w Work, end func()) {
	x.chStartWork <- runWork{w, ordinal, end}
}

func (x Runner) Interrupted() bool {
	ch := make(chan bool)
	x.chGetInterrupted <- ch
	return <-ch
}

func (x Runner) Interrupt() {
	x.chInterrupt <- struct{}{}
}

func (x Runner) WriteError(productSerial ankat.ProductSerial, text string) {
	x.WriteLog(productSerial, worklog.Error, text)
}

func (x Runner) WriteLog(productSerial ankat.ProductSerial, level worklog.Level, text string) {
	x.chWriteLog <- worklog.WriteRecord{
		Level:         level,
		Text:          text,
		ProductSerial: productSerial,
	}
}

func (x Runner) WriteLogf(productSerial ankat.ProductSerial, level worklog.Level, format string, a ...interface{}) {
	x.WriteLog(productSerial, level, fmt.Sprintf(format, a...))
}

func (x Runner) Delay(name string, duration time.Duration, backgroundWork func() error) error {
	timer := time.NewTimer(duration)

	x.chDelay <- delayInfo{
		Name:       name,
		DurationMS: duration.Nanoseconds() / 1e6,
		Enabled:    true,
	}
	defer func() {
		x.chDelay <- delayInfo{}
	}()
	x.WriteLogf(0, worklog.Info, "%s %v", name, duration)
	startTime := time.Now()

	for {
		if x.Interrupted() {
			return errorInterrupted
		}

		select {
		case <-timer.C:
			return nil
		case <-x.chDelaySkipped:
			x.WriteLogf(0, worklog.Warning, "%s: %v: прервано пользователем %v: ", name, duration, time.Since(startTime))
			return nil
		default:
			if backgroundWork != nil {
				if err := backgroundWork(); err != nil {
					return err
				}
			}
		}
	}
}

func (x Runner) WorkDelay(name string, getDuration func() time.Duration, backgroundWork func() error) Work {
	return Work{
		Name: name,
		Action: func() error {
			return x.Delay(name, getDuration(), backgroundWork)
		},
	}
}

//func notifyEnd(what string, err error) (t coloredText) {
//	if err == nil {
//		t.Text = fmt.Sprintf("%s: выполнено без ошибок", what)
//		t.Color = "clNavy"
//	} else {
//		t.Text = fmt.Sprintf("%s: %s", what, err)
//		t.Color = "clRed"
//	}
//	return
//}

func mustUnmarshalJson(b []byte, v interface{}) {
	if err := json.Unmarshal(b, v); err != nil {
		panic(err.Error() + ": " + string(b))
	}
}

func fmtErr(e error) string {
	if e != nil {
		return e.Error()
	}
	return ""
}
