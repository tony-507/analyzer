package common

type Plugin struct {
	setCallback   func(RequestHandler)
	setParameter  func(string)
	setResource   func(*ResourceLoader)
	startSequence func()
	deliverUnit   func(CmUnit)
	deliverStatus func(CmUnit)
	fetchUnit     func() CmUnit
	endSequence   func()
	name          string
	isRoot        bool
}

func CreatePlugin(
	name string,
	isRoot bool,
	setCallback func(RequestHandler),
	setParameter func(string),
	setResource func(*ResourceLoader),
	startSequence func(),
	deliverUnit func(CmUnit),
	deliverStatus func(CmUnit),
	fetchUnit func() CmUnit,
	endSequence func(),
) Plugin {
	return Plugin{
		setCallback:   setCallback,
		setParameter:  setParameter,
		setResource:   setResource,
		startSequence: startSequence,
		deliverUnit:   deliverUnit,
		deliverStatus: deliverStatus,
		fetchUnit:     fetchUnit,
		endSequence:   endSequence,
		name:          name,
		isRoot:        isRoot,
	}
}

// Method protection
func (pg *Plugin) SetCallback(h RequestHandler) {
	pg.setCallback(h)
}

func (pg *Plugin) SetParameter(m_parameter string) {
	pg.setParameter(m_parameter)
}

func (pg *Plugin) SetResource(loader *ResourceLoader) {
	pg.setResource(loader)
}

func (pg *Plugin) StartSequence() {
	pg.startSequence()
}

func (pg *Plugin) DeliverUnit(unit CmUnit) {
	pg.deliverUnit(unit)
}

func (pg *Plugin) DeliverStatus(unit CmUnit) {
	pg.deliverStatus(unit)
}

func (pg *Plugin) FetchUnit() CmUnit {
	return pg.fetchUnit()
}

func (pg *Plugin) EndSequence() {
	pg.endSequence()
}

func (pg *Plugin) IsRoot() bool {
	return pg.isRoot
}

func (pg *Plugin) Name() string {
	return pg.name
}
