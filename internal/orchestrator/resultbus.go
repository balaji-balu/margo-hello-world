package orchestrator

type WorkResult struct {
    Node   string
    Status string // "success" or "fail"
    Error  error
}

type ResultBus struct {
    Results chan WorkResult
}

func NewResultBus() *ResultBus {
    return &ResultBus{Results: make(chan WorkResult, 10)}
}

func (rb *ResultBus) Publish(node string, status string, err error) {
    rb.Results <- WorkResult{Node: node, Status: status, Error: err}
}
