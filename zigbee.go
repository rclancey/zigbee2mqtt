package zigbee2mqtt

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

type Target struct {
	Endpoint int `json:"endpoint"`
	IEEEAddress string `json:"ieee_address"`
	Type string `json:"type"`
}

type Binding struct {
	Cluster string `json:"cluster"`
	Target Target `json:"target"`
}

type Cluster struct {
	Input []string `json:"input"`
	Output []string `json:"output"`
}

type Reporting struct {
	Attribute string `json:"attribute"`
	Cluster string `json:"cluster"`
	MaximumReportInterval int `json:"maximum_report_interval"`
	MinimumReportInterval int `json:"minimum_report_interval"`
	ReportableChange int `json:"reportable_change"`
}

type Endpoint struct {
	Bindings []Binding `json:"bindings"`
	Clusters []Cluster `json:"clusters"`
	ConfiguredReportings []Reporting `json:"configured_reportings"`
	Scenes []Scene `json:"scenes"`
}

type Feature struct {
	Access int `json:"access"`
	Category string `json:"category"`
	Description string `json:"description"`
	Label string `json:"label"`
	Name string `json:"name"`
	Property string `json:"property"`
	Type string `json:"type"`
	Unit string `json:"unit"`
	ValueOn any `json:"value_on"`
	ValueOff any `json:"value_off"`
	ValueMin *float64 `json:"value_min"`
	ValueMax *float64 `json:"value_max"`
	Values []string `json:"values"`
	Features []*Feature `json:"features"`
	ItemType *Feature `json:"item_type"`
}

func (feat *Feature) Readable() bool {
	return feat.Access & 1 != 0
}

func (feat *Feature) Setable() bool {
	return feat.Access & 2 != 0
}

func (feat *Feature) Getable() bool {
	return feat.Access & 4 != 0
}

/*
func (feat *Feature) Set(val any) (any, error) {
	switch feat.Type {
	case "binary":
		switch v := val.(type) {
		case bool:
			if v {
				if feat.ValueOn == nil {
					return true, nil
				}
				return feat.ValueOn, nil
			}
			if feat.ValueOff == nil {
				return false, nil
			}
			return feat.ValueOff, nil
		case string:
			if feat.ValueOn != nil {
				s, ok := feat.ValueOn.(string)
				if ok && s == v {
					return v, nil
				}
			}
			if feat.ValueOff != nil {
				s, ok := feat.ValueOff.(string)
				if ok && s == v {
					return v, nil
				}
			}
			if feat.ValueOn != nil && feat.ValueOff != nil {
				
			strVals := map[string]bool{
				"on": true,
				"true": true,
				"yes": true,
				"off": false,
				"false": false,
				"no": false,
			}
			b, ok := strVals[strings.ToLower(v)]
			if ok {
				return b, nil
			}
			return nil, errors.New("invalid value")
		case float64:
			
		case int:
		case int64:
		case int32:
		case int16:
		case int8:
		case uint:
		case uint64:
		case uint32:
		case uint16:
		case uint8:
		case float32:
		}
	case "numeric":
		switch v := val.(type) {
		case float64:
		case int:
		case int64:
		case int32:
		case int16:
		case int8:
		case uint:
		case uint64:
		case uint32:
		case uint16:
		case uint8:
		case float32:
		}
	case "enum":
		switch v := val.(type) {
		case string:
		case int:
		}
	case "text":
		switch v := val.(type) {
		case string:
		case fmt.Stringer:
		case encoding.TextMarshaler:
		}
	case "list":
		// complicated
	default:
		// composite, specific
	}
}

func (feat *Feature) WithValue(val any) *Feature {
}
*/

type Definition struct {
	Vendor string `json:"vendor"`
	Model string `json:"model"`
	Description string `json:"description"`
	SupportsOTA bool `json:"supports_ota"`
	Exposes []*Feature `json:"exposes"`
	Options []*Feature `json:"options"`
}

type Scene struct {
}

type DeviceIdentifier struct {
	IEEEAddress string `json:"ieee_address"`
	FriendlyName string `json:"friendly_name"`
}

type Device struct {
	DeviceIdentifier
	Type string `json:"type"`
	Manufacturer string `json:"manufacturer"`
	ModelID string `json:"model_id"`
	NetworkAddress int `json:"network_address"`
	Description string `json:"description"`
	Endpoints map[string]Endpoint `json:"endpoints"`
	Definition *Definition `json:"definition"`
	PowerSource string `json:"power_source"`
	DateCode string `json:"date_code"`
	Scenes []Scene `json:"scene"`
	Interviewing bool `json:"interviewing"`
	InterviewCompleted bool `json:"interview_completed"`
	Supported bool `json:"supported"`
	Disabled bool `json:"disabled"`
	client mqtt.Client
}

/*
func (dev Device) Set(key string, value any) error {
	if dev.Definition == nil {
		return errors.New("no definition for device")
	}
	for _, opt := range dev.Definition.Options {
		if opt.Property == key {
			switch opt.Type {
			case "binary":
			case "numeric":
			case 
}
*/

type State struct {
	State string `json:"state"`
}

type LogMessage struct {
	Level string `json:"level"`
	Message string `json:"message"`
}

type EventData struct {
	FriendlyName string `json:"friendly_name"`
	IEEEAddress string `json:"ieee_address"`
	Status *string `json:"status"`
	Supported *bool `json:"supported"`
	Definition *Definition `json:"definition"`
}

type Event struct {
	Type string
	Data EventData
}

type DeviceHandler func(dev *Device) error
type TelemHandler func(dev *Device, telem json.RawMessage) error
type BridgeStateHandler func(status string) error
type LoggingHandler func(level, message string) error
type EventHandler func(evt *Event) error

type TypedMessageHandler[T any] func(topic string, msg T) error

func Handler[T any](handler TypedMessageHandler[T]) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		var obj T
		err := json.Unmarshal(message.Payload(), &obj)
		if err != nil {
			log.Printf("error unmarshaling %s message into %T: %s", message.Topic(), obj, err)
			log.Println(string(message.Payload()))
			return
		}
		err = handler(message.Topic(), obj)
		if err != nil {
			log.Printf("error handling %s message: %s", message.Topic(), err)
			return
		}
	}
}

func TypedTelemHandler[T any](handler func(*Device, T) error) TelemHandler {
	return func(dev *Device, telem json.RawMessage) error {
		var obj T
		err := json.Unmarshal(telem, &obj)
		if err != nil {
			return err
		}
		return handler(dev, obj)
	}
}

type Request struct {
	Data any
	Transaction int64
}

func (req Request) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(req.Data)
	if err != nil {
		return nil, err
	}
	obj := map[string]any{}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	obj["transaction"] = req.Transaction
	return json.Marshal(obj)
}

type Response struct {
	Data json.RawMessage `json:"data"`
	Transaction int64 `json:"transaction"`
	Status *string `json:"status"`
}

type DeviceUpdate struct {
	Device *Device
	Update map[string]any
}

type Zigbee struct {
	client mqtt.Client
	devices []*Device
	deviceSubs map[string]bool
	bridgeState string
	deviceHandler DeviceHandler
	bridgeStateHandler BridgeStateHandler
	loggingHandler LoggingHandler
	eventHandler EventHandler
	transactionMutex *sync.Mutex
	nextTransactionId *atomic.Int64
	pendingTransactions map[int64]chan Response
	updateChan chan DeviceUpdate
}

func NewZigbee(client mqtt.Client) *Zigbee {
	z := &Zigbee{
		client: client,
		deviceSubs: map[string]bool{},
		transactionMutex: &sync.Mutex{},
		nextTransactionId: &atomic.Int64{},
		pendingTransactions: map[int64]chan Response{},
		updateChan: make(chan DeviceUpdate, 1000),
	}
	z.nextTransactionId.Store(rand.Int63())
	client.Subscribe("zigbee2mqtt/bridge/devices", 0, Handler(func(topic string, msg []*Device) error {
		z.devices = msg
		active := map[string]bool{}
		for i := range msg {
			dev := msg[i]
			if !z.deviceSubs[dev.FriendlyName] {
				log.Printf("subscribing to device %s (%s)", dev.FriendlyName, dev.IEEEAddress)
				z.client.Subscribe(fmt.Sprintf("zigbee2mqtt/%s", dev.FriendlyName), 0, Handler(func(topic string, msg map[string]any) error {
					z.updateChan <- DeviceUpdate{dev, msg}
					return nil
				}))
			}
			active[dev.FriendlyName] = true
		}
		if z.deviceHandler != nil {
			for _, dev := range msg {
				z.deviceHandler(dev)
			}
		}
		return nil
	}))
	client.Subscribe("zigbee2mqtt/bridge/state", 0, Handler(func(topic string, msg State) error {
		z.bridgeState = msg.State
		if z.bridgeStateHandler != nil {
			z.bridgeStateHandler(msg.State)
		}
		return nil
	}))
	client.Subscribe("zigbee2mqtt/bridge/logging", 0, Handler(func(topic string, msg LogMessage) error {
		if z.loggingHandler != nil {
			z.loggingHandler(msg.Level, msg.Message)
		}
		return nil
	}))
	client.Subscribe("zigbee/bridge/event", 0, Handler(func(topic string, msg Event) error {
		if z.eventHandler != nil {
			z.eventHandler(&msg)
		}
		return nil
	}))
	client.Subscribe("zigbee/bridge/request/#", 0, Handler(func(topic string, msg Response) error {
		if msg.Status != nil {
			z.finishTransaction(msg.Transaction, &msg)
		}
		return nil
	}))
	return z
}

func (z *Zigbee) Updates() <-chan DeviceUpdate {
	return z.updateChan
}

func (z *Zigbee) SetOption(device string, obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	z.client.Publish(fmt.Sprintf("zigbee2mqtt/%s/set", device), 2, false, data)
	return nil
}

func (z *Zigbee) createTransaction() (int64, chan Response) {
	id := z.nextTransactionId.Add(1)
	ch := make(chan Response)
	z.transactionMutex.Lock()
	z.pendingTransactions[id] = ch
	z.transactionMutex.Unlock()
	return id, ch
}

func (z *Zigbee) finishTransaction(id int64, msg *Response) {
	z.transactionMutex.Lock()
	ch, ok := z.pendingTransactions[id]
	if ok {
		delete(z.pendingTransactions, id)
	}
	z.transactionMutex.Unlock()
	if ok && msg != nil {
		ch <- *msg
		close(ch)
	}
}

func execTransaction[T any](z *Zigbee, command string, data any) (T, error) {
	transactionId, ch := z.createTransaction()
	req := &Request{Data: data, Transaction: transactionId}
	payload, err := json.Marshal(req)
	var out T
	if err != nil {
		go z.finishTransaction(transactionId, nil)
		return out, err
	}
	z.client.Publish(fmt.Sprintf("zigbee2mqtt/bridge/request/%s", command), 2, false, payload)
	resp := <-ch
	if resp.Status != nil && *resp.Status != "ok" {
		json.Unmarshal(resp.Data, &out)
		return out, errors.New(*resp.Status)
	}
	err = json.Unmarshal(resp.Data, &out)
	return out, err
}

type PermitJoinCommand struct {
	Value bool `json:"value"`
	Device string `json:"device,omitempty"`
	Time int `json:"time,omitempty"`
}

func (z *Zigbee) PermitJoin() error {
	_, err := execTransaction[PermitJoinCommand](z, "permit_join", PermitJoinCommand{Value: true})
	return err
}

func (z *Zigbee) ProhibitJoin() error {
	_, err := execTransaction[PermitJoinCommand](z, "permit_join", PermitJoinCommand{Value: false})
	return err
}

func (z *Zigbee) PermitJoinFor(dur time.Duration) error {
	_, err :=  execTransaction[PermitJoinCommand](z, "permit_join", PermitJoinCommand{Value: true, Time: int(math.Round(dur.Seconds()))})
	return err
}

type HealthCheck struct {
	Healthy bool `json:"healthy"`
}

func (z *Zigbee) HealthCheck() (HealthCheck, error) {
	return execTransaction[HealthCheck](z, "health_check", map[string]any{})
}

type CoordinatorCheck struct {
	MissingRouters []DeviceIdentifier `json:"missing_routers"`
}

func (z *Zigbee) CoordinatorCheck() (CoordinatorCheck, error) {
	 return execTransaction[CoordinatorCheck](z, "coordinator_check", map[string]any{})
}

func (z *Zigbee) Restart() error {
	_, err := execTransaction[map[string]any](z, "restart", map[string]any{})
	return err
}

type InstallCode struct {
	Value string `json:"value"`
}

func (z *Zigbee) InstallCode(code string) error {
	_, err := execTransaction[InstallCode](z, "install_code/add", InstallCode{Value: code})
	return err
}

type RemoveDevice struct {
	ID string `json:"id"`
	Block *bool `json:"block,omitempty"`
	Force *bool `json:"force,omitempty"`
}

func (z *Zigbee) RemoveDevice(id string) error {
	_, err := execTransaction[RemoveDevice](z, "device/remove", RemoveDevice{ID: id})
	return err
}

type InterviewDevice struct {
	ID string `json:"id"`
}

func (z *Zigbee) InterviewDevice(id string) error {
	_, err := execTransaction[InterviewDevice](z, "device/interview", InterviewDevice{ID: id})
	return err
}

type RenameDevice struct {
	From string `json:"from"`
	To string `json:"to"`
	HomeAssistantRename bool `json:"homeassistant_rename,omitempty"`
}

func (z *Zigbee) RenameDevice(id, name string) error {
	_, err := execTransaction[RenameDevice](z, "device/rename", RenameDevice{From: id, To: name})
	return err
}

type ConfigureReporting struct {
	ID string `json:"id"`
	Cluster string `json:"cluster"`
	Attribute string `json:"attribute"`
	MinimumReportInterval int `json:"minimum_report_interval"`
	MaximumReportInterval int `json:"maximum_report_interval"`
	ReportableChange float64 `json:"reportable_change"`
}

func (z *Zigbee) ConfigureReporting(id, cluster, attribute string, minInterval, maxInterval time.Duration, thresh float64) error {
	_, err := execTransaction[ConfigureReporting](z, "device/configure_reporting", ConfigureReporting{
		ID: id,
		Cluster: cluster,
		Attribute: attribute,
		MinimumReportInterval: int(math.Round(minInterval.Seconds())),
		MaximumReportInterval: int(math.Round(maxInterval.Seconds())),
		ReportableChange: thresh,
	})
	return err
}

type AddGroup struct {
	ID int `json:"id"`
	FriendlyName string `json:"friendly_name"`
}

func (z *Zigbee) AddGroup(id int, name string) error {
	_, err := execTransaction[AddGroup](z, "group/add", AddGroup{ID: id, FriendlyName: name})
	return err
}

type RemoveGroup struct {
	ID any `json:"id"`
}

func (z *Zigbee) RemoveGroup(id any) error {
	_, err := execTransaction[RemoveGroup](z, "group/remove", RemoveGroup{ID: id})
	return err
}

type RenameGroup struct {
	From any `json:"from"`
	To string `json:"to"`
}

func (z *Zigbee) RenameGroup(id any, name string) error {
	_, err := execTransaction[RenameGroup](z, "group/rename", RenameGroup{From: id, To: name})
	return err
}

type GroupOptions struct {
	Retain *bool `json:"retain,omitempty"`
	Transition *int `json:"transition,omitempty"`
	Optimistic *bool `json:"optimistic,omitempty"`
	OffState *string `json:"off_state,omitempty"`
}

type SetGroupOptions struct {
	ID any `json:"id"`
	Options *GroupOptions `json:"options,omitempty"`
	From *GroupOptions `json:"from,omitempty"`
	To *GroupOptions `json:"to,omitempty"`
}

func (z *Zigbee) SetGroupOptions(id any, opts GroupOptions) (SetGroupOptions, error) {
	return execTransaction[SetGroupOptions](z, "group/options", SetGroupOptions{ID: id, Options: &opts})
}

type GroupMembers struct {
	Group string `json:"group,omitempty"`
	Device string `json:"device"`
}

func (z *Zigbee) AddGroupMembers(group, device string) error {
	_, err := execTransaction[GroupMembers](z, "group/members/add", GroupMembers{Group: group, Device: device})
	return err
}

func (z *Zigbee) RemoveGroupMembers(group, device string) error {
	_, err := execTransaction[GroupMembers](z, "group/members/remove", GroupMembers{Group: group, Device: device})
	return err
}

func (z *Zigbee) RemoveAllGroupMembers(device string) error {
	_, err := execTransaction[GroupMembers](z, "group/members/remove", GroupMembers{Device: device})
	return err
}

