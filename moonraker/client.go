package moonraker

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	c                *websocket.Conn
	cMu              sync.Mutex
	callback         map[int64]clientCallback
	callbackRegister chan clientCallbackRegister
	stop             chan struct{}
	id               int64
}

type clientCallback struct {
	response chan json.RawMessage
	once     bool
}

type clientCallbackRegister struct {
	id       int64
	response chan json.RawMessage
	once     bool
}

func New(addr string, apiKey ...string) (*Client, error) {
	var h http.Header
	if len(apiKey) > 0 {
		h.Add("X-API-KEY", apiKey[0])
	}

	n, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/websocket", h)
	if err != nil {
		return nil, err
	}

	c := &Client{
		c:                n,
		callback:         make(map[int64]clientCallback),
		callbackRegister: make(chan clientCallbackRegister, 8),
		stop:             make(chan struct{}),
	}
	go c.run()

	return c, nil
}

func (c *Client) Request(method string, params, value interface{}) error {
	if params == nil {
		params = make(map[string]interface{})
	}

	rawParams, err := json.Marshal(params)
	if err != nil {
		return err
	}

	request := jsonRPCRequest{
		Version: "2.0",
		Method:  method,
		Params:  rawParams,
		ID:      atomic.AddInt64(&c.id, 1),
	}

	var message []byte
	if message, err = json.Marshal(request); err != nil {
		return err
	}

	if log.StandardLogger().Level < log.TraceLevel {
		log.WithFields(log.Fields{
			"method": method,
		}).Debug("JSON-RPC request")
	} else {
		log.WithFields(log.Fields{
			"method": method,
			"params": string(rawParams),
			"id":     request.ID,
		}).Trace("JSON-RPC request")
	}

	c.cMu.Lock()
	err = c.c.WriteMessage(websocket.TextMessage, message)
	c.cMu.Unlock()
	if err != nil {
		return err
	}

	if value != nil {
		// Wait for answer
		response := make(chan json.RawMessage, 1)
		c.callbackRegister <- clientCallbackRegister{
			id:       request.ID,
			response: response,
			once:     true,
		}
		defer close(response)
		return json.Unmarshal(<-response, value)
	}

	return nil
}

type TemperatureStore struct {
	Temperatures []float64 `json:"temperatures"`
	Targets      []float64 `json:"targets"`
	Powers       []float64 `json:"powers"`
}

func (c *Client) TemperatureStores() (store map[string]TemperatureStore, err error) {
	store = make(map[string]TemperatureStore)
	err = c.Request("server.temperature_store", nil, &store)
	return
}

type PrinterInfo struct {
	State           string `json:"state"`
	StateMessage    string `json:"state_message"`
	Hostname        string `json:"hostname"`
	SoftwareVersion string `json:"software_version"`
	CPUInfo         string `json:"cpu_info"`
	KlipperPath     string `json:"klipper_path"`
	PythonPath      string `json:"python_path"`
	LogFile         string `json:"log_file"`
	ConfigFile      string `json:"config_file"`
}

func (c *Client) PrinterInfo() (info PrinterInfo, err error) {
	err = c.Request("printer.info", nil, &info)
	return
}

func (c *Client) EmergencyStop() error {
	return c.Request("printer.emergency_stop", nil, nil)
}

func (c *Client) Restart() error {
	return c.Request("printer.restart", nil, nil)
}

func (c *Client) RestartFirmware() error {
	return c.Request("printer.firmware_restart", nil, nil)
}

type Query struct {
	Objects map[string][]string `json:"objects"`
}

func NewQuery(object string) *Query {
	return &Query{
		Objects: map[string][]string{object: nil},
	}
}

type QueryResult struct {
	Status map[string]json.RawMessage `json:"status"`
}

func NewQueryResult() *QueryResult {
	return &QueryResult{
		Status: make(map[string]json.RawMessage),
	}
}

func (c *Client) queryOne(object string, value interface{}) (err error) {
	log.WithField("object", object).Debug("")
	var (
		query  = NewQuery(object)
		result = NewQueryResult()
	)
	if err = c.Request("printer.objects.query", query, result); err == nil {
		err = json.Unmarshal(result.Status[object], value)
	}
	return
}

type ExtruderStatus struct {
	Temperature     float64 `json:"temperature"`
	Target          float64 `json:"target"`
	Power           float64 `json:"power"`
	PressureAdvance float64 `json:"pressure_advance"`
	SmoothTime      float64 `json:"smooth_time"`
}

func (c *Client) ExtruderStatus(toolhead int) (status ExtruderStatus, err error) {
	name := "extruder"
	if toolhead > 0 {
		name += strconv.FormatInt(int64(toolhead), 10)
	}
	if err = c.queryOne(name, &status); err != nil {
		log.Println("moonraker: extruder status:", err)
	}
	return
}

type GCodeStatus struct {
	SpeedFactor         float64    `json:"speed_factor"`
	Speed               float64    `json:"speed"`
	ExtrudeFactor       float64    `json:"extrude_factor"`
	AbsoluteCoordinates bool       `json:"absolute_coordinates"`
	HomingOrigin        [4]float64 `json:"homing_origin"`
	Position            [4]float64 `json:"position"`
	GCodePosition       [4]float64 `json:"gcode_position"`
}

func (c *Client) GCodeStatus() (status GCodeStatus, err error) {
	err = c.queryOne("gcode_move", &status)
	return
}

type ToolheadStatus struct {
	HomedAxes            string     `json:"homed_axes"`
	PrintTime            float64    `json:"print_time"`
	EstimatedPrintTime   float64    `json:"estimated_print_time"`
	Extruder             string     `json:"extruder"`
	Position             [4]float64 `json:"position"`
	MaxVelocity          float64    `json:"max_velocity"`
	MaxAccel             float64    `json:"max_accel"`
	MaxAccelToDecel      float64    `json:"max_accel_to_decel"`
	SquareCornerVelocity float64    `json:"square_corner_velocity"`
}

func (c *Client) ToolheadStatus() (status ToolheadStatus, err error) {
	err = c.queryOne("toolhead", &status)
	return
}

type HeaterBedStatus struct {
	Temperature float64 `json:"temperature"`
	Target      float64 `json:"target"`
	Power       float64 `json:"power"`
}

func (c *Client) HeaterBedStatus() (status HeaterBedStatus, err error) {
	err = c.queryOne("heater_bed", &status)
	return
}

type FanStatus struct {
	Speed float64 `json:"speed"`
	RPM   float64 `json:"rpm"`
}

func (c *Client) FanStatus() (status FanStatus, err error) {
	err = c.queryOne("fan", &status)
	return
}

func (c *Client) run() {
	var received = make(chan *jsonRPCResponse, 8)

	go func(received chan<- *jsonRPCResponse) {
		for {
			var r = new(jsonRPCResponse)
			if err := c.c.ReadJSON(r); err != nil {
				//log.Println("moonraker: JSON receive error: %w", err)
				log.WithError(err).Error("failed to parse Moonraker JSON-RPC response")
				return
			} else {
				received <- r
			}
		}
	}(received)

	// Announce as Klipper online
	dispatch(eventOnline, c, nil)

	for {
		select {
		case r := <-c.callbackRegister:
			c.callback[r.id] = clientCallback{
				response: r.response,
				once:     r.once,
			}

		case r := <-received:
			logger := log.WithFields(log.Fields{
				"method": r.Method,
				"id":     r.ID,
				"result": string(r.Result),
				"params": string(r.Params),
			})
			logger.Trace("JSON-RPC response")
			if r.ID == 0 {
				// Moonraker stats, ignored for now
			} else {
				//log.Println("moonraker:", r.Method+":", string(r.Result))
				if callback, ok := c.callback[r.ID]; ok {
					callback.response <- r.Result
					if callback.once {
						delete(c.callback, r.ID)
					}
				} else {
					logger.Warn("unhandled Moonraker API response!")
				}
			}

		case <-c.stop:
			return
		}
	}
}

type jsonRPCRequest struct {
	Version string          `json:"jsonrpc"` // Always 2.0
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      int64           `json:"id"`
}

type jsonRPCResponse struct {
	jsonRPCRequest
	Result json.RawMessage `json:"result"`
}
