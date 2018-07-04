package uiworks

import (
	"encoding/json"
	"fmt"
	"github.com/fpawel/procmq"
	"time"

	"github.com/fpawel/ankat/data/dataworks"
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
	chSubscribeInterrupted   chan chan struct{}
	chUnsubscribeInterrupted chan chan struct{}
	chDelay                  chan delayInfo
	chDelaySkipped           chan struct{}
	chWriteLog               chan dataworks.WriteRecord
	chStartWork              chan runWork
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
		chSubscribeInterrupted:   make(chan chan struct{}),
		chUnsubscribeInterrupted: make(chan chan struct{}),
		chDelay:                  make(chan delayInfo),
		chDelaySkipped:           make(chan struct{}),
		chWriteLog:               make(chan dataworks.WriteRecord),
	}

	processMQ.Handle("CURRENT_WORK_START", func(bytes []byte) {
		var payload notifyWork
		mustUnmarshalJson(bytes, &payload)
		x.chStartMainTask <- payload
	})
	processMQ.Handle("CURRENT_WORK_STOP", func([]byte) {
		x.chInterrupt <- struct{}{}
	})
	processMQ.Handle("SKIP_DELAY", func([]byte) {
		x.chDelaySkipped <- struct{}{}
	})

	return x
}

func (x Runner) Close() {
	x.chClose <- struct{}{}
}

func (x Runner) Run(dbLog, dbConfig *sqlx.DB, mainTask *Task) {

	rootTask := mainTask
	started := false
	interrupted := false
	closed := false
	interruptNotifier := &interruptNotifier{}
	workLogWritten := false

	var currentRunTask *Task

	writeLog := func(m dataworks.WriteRecord) {
		if !workLogWritten {
			dataworks.AddRootWork(dbLog, rootTask.name)
			workLogWritten = true
		}
		m.Works = currentRunTask.ParentKeys()
		x.delphiApp.Send("CURRENT_WORK_MESSAGE", dataworks.Write(dbLog, m))
	}

	for {
		select {

		case <-x.chClose:

			if started {
				closed = true
				interrupted = true
				interruptNotifier.Notify()
			} else {
				return
			}

		case <-x.chInterrupt:
			interrupted = true
			interruptNotifier.Notify()

		case m := <-x.chWriteLog:
			writeLog(m)

		case v := <-x.chDelay:
			x.delphiApp.Send("DELAY", v)

		case ch := <-x.chSubscribeInterrupted:
			interruptNotifier.Subscribe(ch)

		case ch := <-x.chUnsubscribeInterrupted:
			interruptNotifier.Unsubscribe(ch)

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
			if !m.Checked(dbConfig) {
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
			interruptNotifier.Notify()
			currentRunTask = nil
			x.delphiApp.Send("END_WORK", struct {
				Name, Error string
			}{v.Name, fmtErr(v.Error)})

		case <-x.chStop:
			if started {
				interrupted = true
				interruptNotifier.Notify()
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

func (x Runner) SubscribeInterrupted(ch chan struct{}, subscribe bool) {
	if subscribe {
		x.chSubscribeInterrupted <- ch
	} else {
		x.chUnsubscribeInterrupted <- ch
	}
}

func (x Runner) WriteLog(productSerial int, level dataworks.Level, text string) {
	x.chWriteLog <- dataworks.WriteRecord{
		Level:         level,
		Text:          text,
		ProductSerial: productSerial,
	}
}

func (x Runner) WriteLogf(productSerial int, level dataworks.Level, format string, a ...interface{}) {
	x.WriteLog(productSerial, level, fmt.Sprintf(format, a...))
}

func (x Runner) Delay(name string, duration time.Duration, backgroundWork func() error) error {
	timer := time.NewTimer(duration)
	chInterrupted := make(chan struct{})
	x.SubscribeInterrupted(chInterrupted, true)
	defer x.SubscribeInterrupted(chInterrupted, false)

	x.chDelay <- delayInfo{
		Name:       name,
		DurationMS: duration.Nanoseconds() / 1e6,
		Enabled:    true,
	}
	defer func() {
		x.chDelay <- delayInfo{}
	}()
	x.WriteLog(0, dataworks.Debug, fmt.Sprintf("Задержка %v", duration))

	for {
		select {
		case <-timer.C:
			return nil
		case <-chInterrupted:
			return errorInterrupted
		case <-x.chDelaySkipped:
			x.WriteLog(0, dataworks.Warning, "задержка прервана")
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

type interruptNotifier struct {
	chs []chan struct{}
}

func (x *interruptNotifier) Notify() {
	for _, ch := range x.chs {
		ch <- struct{}{}
	}
	x.chs = nil
}

func (x *interruptNotifier) Subscribe(ch chan struct{}) {
	x.chs = append(x.chs, ch)
}

func (x *interruptNotifier) Unsubscribe(ch chan struct{}) {
	for i, cha := range x.chs {
		if cha == ch {
			x.chs[i] = x.chs[len(x.chs)-1]
			x.chs[len(x.chs)-1] = nil
			x.chs = x.chs[:len(x.chs)-1]
		}
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
