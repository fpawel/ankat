package uiworks

type Work struct {
	Name     string
	Children []Work
	Action   func() error
}

func (x *Work) AddChild(w Work) {
	x.Children = append(x.Children, w)
}

func (x *Work) AddChildren(children ...Work) {
	x.Children = append(x.Children, children...)
}

func S(name string, action Action) Work {
	return Work{
		Name:   name,
		Action: action,
	}
}

func L(name string, children ...Work) Work {
	return Work{
		Name:     name,
		Children: children,
	}
}

func (x Work) Check() {
	if len(x.Children) == 0 && x.Action == nil {
		panic("uiworks.Task: нет операции и нет потомков")
	}
	if len(x.Children) != 0 && x.Action != nil {
		panic("uiworks.Task: есть операция и есть потомки")
	}
}

func (x Work) Task() *Task {
	x.Check()
	m := &Task{
		name:   x.Name,
		action: x.Action,
	}

	for _, o := range x.Children {
		om := o.Task()
		om.parent = m
		m.children = append(m.children, om)
	}
	if m.parent == nil {
		m.enumDescendants(&m.descendants)
		for i, y := range m.descendants {
			y.ordinal = i
		}
	}

	return m
}
