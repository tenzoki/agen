package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
	"github.com/tenzoki/agen/omni/public/omnistore"
)

// PEVKnowledgeStore records PEV workflow history in OmniStore graph
type PEVKnowledgeStore struct {
	agent.DefaultAgentRunner
	store      omnistore.OmniStore
	dataPath   string
	baseAgent  *agent.BaseAgent
}

// Request record stored in knowledge base
type RequestRecord struct {
	ID              string                 `json:"id"`
	Content         string                 `json:"content"`
	Context         map[string]interface{} `json:"context"`
	Timestamp       int64                  `json:"timestamp"`
	Completed       bool                   `json:"completed"`
	FinalIteration  int                    `json:"final_iteration,omitempty"`
	SuccessfulPlanID string                 `json:"successful_plan_id,omitempty"`
}

// PlanRecord stored in knowledge base
type PlanRecord struct {
	ID          string                   `json:"id"`
	RequestID   string                   `json:"request_id"`
	Goal        string                   `json:"goal"`
	Steps       []map[string]interface{} `json:"steps"`
	Timestamp   int64                    `json:"timestamp"`
	Iteration   int                      `json:"iteration"`
	Successful  bool                     `json:"successful,omitempty"`
}

// VerificationRecord stored in knowledge base
type VerificationRecord struct {
	ID           string `json:"id"`
	RequestID    string `json:"request_id"`
	PlanID       string `json:"plan_id"`
	GoalAchieved bool   `json:"goal_achieved"`
	Iteration    int    `json:"iteration"`
	IssueCount   int    `json:"issue_count"`
	Timestamp    int64  `json:"timestamp"`
}

func NewPEVKnowledgeStore() *PEVKnowledgeStore {
	return &PEVKnowledgeStore{}
}

func (k *PEVKnowledgeStore) Init(base *agent.BaseAgent) error {
	k.baseAgent = base

	// Get data path from config
	k.dataPath = base.GetConfigString("data_path", "./data/pev-knowledge")

	// Initialize OmniStore
	store, err := omnistore.NewOmniStoreWithDefaults(k.dataPath)
	if err != nil {
		return fmt.Errorf("failed to initialize OmniStore: %w", err)
	}
	k.store = store

	base.LogInfo("PEV Knowledge Store initialized (data path: %s)", k.dataPath)
	return nil
}

func (k *PEVKnowledgeStore) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	base.LogDebug("Knowledge store received message type: %s", msg.Type)

	switch msg.Type {
	case "user_request":
		return k.recordRequest(msg, base)
	case "execution_plan":
		return k.recordPlan(msg, base)
	case "execution_results":
		return k.recordExecution(msg, base)
	case "verification_report":
		return k.recordVerification(msg, base)
	case "query_similar":
		return k.querySimilarRequests(msg, base)
	default:
		// Ignore unknown message types
		return nil, nil
	}
}

func (k *PEVKnowledgeStore) Cleanup(base *agent.BaseAgent) {
	if k.store != nil {
		k.store.Close()
	}
	base.LogInfo("PEV Knowledge Store cleanup complete")
}

// recordRequest stores a user request in KV store
func (k *PEVKnowledgeStore) recordRequest(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid payload format")
	}

	requestID, _ := payload["id"].(string)
	content, _ := payload["content"].(string)
	context, _ := payload["context"].(map[string]interface{})

	// Create request record
	record := RequestRecord{
		ID:        requestID,
		Content:   content,
		Context:   context,
		Timestamp: time.Now().Unix(),
		Completed: false,
	}

	// Serialize to JSON
	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Store in KV with key: request:{id}
	key := fmt.Sprintf("request:%s", requestID)
	if err := k.store.KV().Set(key, data); err != nil {
		base.LogError("Failed to store request: %v", err)
		return nil, err
	}

	base.LogInfo("Recorded request: %s", requestID)

	// Also index for search
	k.store.Search().IndexDocument(requestID, content, map[string]interface{}{
		"type":      "request",
		"timestamp": time.Now().Unix(),
	})

	return nil, nil
}

// recordPlan stores an execution plan in KV store
func (k *PEVKnowledgeStore) recordPlan(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	plan, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid plan format")
	}

	planID, _ := plan["id"].(string)
	requestID, _ := plan["request_id"].(string)
	goal, _ := plan["goal"].(string)
	steps, _ := plan["steps"].([]interface{})

	// Convert steps to map slice
	stepMaps := make([]map[string]interface{}, 0, len(steps))
	for _, step := range steps {
		if stepMap, ok := step.(map[string]interface{}); ok {
			stepMaps = append(stepMaps, stepMap)
		}
	}

	// Create plan record
	record := PlanRecord{
		ID:        planID,
		RequestID: requestID,
		Goal:      goal,
		Steps:     stepMaps,
		Timestamp: time.Now().Unix(),
		Iteration: 1, // Will be updated on verification
	}

	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plan: %w", err)
	}

	// Store in KV: plan:{id}
	key := fmt.Sprintf("plan:%s", planID)
	if err := k.store.KV().Set(key, data); err != nil {
		base.LogError("Failed to store plan: %v", err)
		return nil, err
	}

	// Also store index: request_plans:{request_id} → [plan_ids]
	plansKey := fmt.Sprintf("request_plans:%s", requestID)
	existingData, _ := k.store.KV().Get(plansKey)
	var planIDs []string
	if existingData != nil {
		json.Unmarshal(existingData, &planIDs)
	}
	planIDs = append(planIDs, planID)
	planIDsData, _ := json.Marshal(planIDs)
	k.store.KV().Set(plansKey, planIDsData)

	base.LogInfo("Recorded plan: %s with %d steps", planID, len(steps))

	// Index for search
	k.store.Search().IndexDocument(planID, goal, map[string]interface{}{
		"type":       "plan",
		"request_id": requestID,
		"step_count": len(steps),
	})

	return nil, nil
}

// recordExecution - stub for now (optional for phase 8 MVP)
func (k *PEVKnowledgeStore) recordExecution(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	// Skip detailed execution recording in MVP
	return nil, nil
}

// recordVerification updates request completion status
func (k *PEVKnowledgeStore) recordVerification(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	report, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid verification report format")
	}

	requestID, _ := report["request_id"].(string)
	goalAchieved, _ := report["goal_achieved"].(bool)

	if !goalAchieved {
		return nil, nil // Skip if not complete
	}

	// Update request record to mark as completed
	reqKey := fmt.Sprintf("request:%s", requestID)
	data, err := k.store.KV().Get(reqKey)
	if err != nil {
		return nil, err
	}

	var request RequestRecord
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, err
	}

	// Get the successful plan ID
	plansKey := fmt.Sprintf("request_plans:%s", requestID)
	plansData, _ := k.store.KV().Get(plansKey)
	var planIDs []string
	if plansData != nil {
		json.Unmarshal(plansData, &planIDs)
		if len(planIDs) > 0 {
			request.SuccessfulPlanID = planIDs[len(planIDs)-1]
			request.FinalIteration = len(planIDs)
		}
	}

	request.Completed = true

	// Save updated request
	updatedData, _ := json.Marshal(request)
	k.store.KV().Set(reqKey, updatedData)

	base.LogInfo("Marked request %s as completed (iteration %d)", requestID, request.FinalIteration)
	return nil, nil
}

// querySimilarRequests finds similar past requests and their outcomes
func (k *PEVKnowledgeStore) querySimilarRequests(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid query payload")
	}

	query, _ := payload["query"].(string)
	limit, _ := payload["limit"].(int)
	if limit == 0 {
		limit = 5
	}

	// Search for similar requests
	searchResult, err := k.store.Search().Search(query, &omnistore.SearchOptions{
		Size: limit,
		Filters: map[string]interface{}{
			"type": "request",
		},
	})

	if err != nil {
		base.LogError("Failed to search similar requests: %v", err)
		return nil, err
	}

	// For each similar request, get its outcome
	var similarRequests []map[string]interface{}

	for _, hit := range searchResult.Hits {
		requestID := hit.ID

		// Get request from KV
		reqKey := fmt.Sprintf("request:%s", requestID)
		reqData, err := k.store.KV().Get(reqKey)
		if err != nil {
			continue
		}

		var request RequestRecord
		if err := json.Unmarshal(reqData, &request); err != nil {
			continue
		}

		// Only include completed requests
		if !request.Completed {
			continue
		}

		// Get the successful plan
		if request.SuccessfulPlanID == "" {
			continue
		}

		planKey := fmt.Sprintf("plan:%s", request.SuccessfulPlanID)
		planData, err := k.store.KV().Get(planKey)
		if err != nil {
			continue
		}

		var plan PlanRecord
		if err := json.Unmarshal(planData, &plan); err != nil {
			continue
		}

		similarRequests = append(similarRequests, map[string]interface{}{
			"request_id":      requestID,
			"content":         request.Content,
			"similarity":      hit.Score,
			"final_iteration": request.FinalIteration,
			"final_plan_id":   request.SuccessfulPlanID,
			"goal":            plan.Goal,
			"step_count":      len(plan.Steps),
			"success":         true,
		})
	}

	base.LogInfo("Found %d similar successful requests", len(similarRequests))

	// Return results
	response := map[string]interface{}{
		"type":             "similar_requests",
		"query":            query,
		"similar_requests": similarRequests,
		"count":            len(similarRequests),
	}

	return &client.BrokerMessage{
		ID:        fmt.Sprintf("similar-%d", time.Now().UnixNano()),
		Type:      "similar_requests_result",
		Target:    "similar_requests_result",
		Payload:   response,
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}, nil
}

func main() {
	store := NewPEVKnowledgeStore()
	if err := agent.Run(store, "pev-knowledge"); err != nil {
		panic(err)
	}
}
