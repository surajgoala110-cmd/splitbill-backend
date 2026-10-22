package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Member struct { ID string `json:"id"`; Name string `json:"name"` }
type Expense struct {
	ID string `json:"id"`; Description string `json:"description"`; Amount int64 `json:"amount"`
	PaidBy string `json:"paidBy"`; Participants []string `json:"participants"`; CreatedAt string `json:"createdAt"`
}
type Group struct { ID string `json:"id"`; Name string `json:"name"`; Members []Member `json:"members"`; Expenses []Expense `json:"expenses"` }
type Store struct { sync.RWMutex; nextID int64; groups map[string]*Group }
func newStore() *Store { return &Store{groups: make(map[string]*Group)} }
func (s *Store) id(prefix string) string { s.nextID++; return fmt.Sprintf("%s_%d", prefix, s.nextID) }

type createGroupRequest struct { Name string `json:"name"` }
type createMemberRequest struct { Name string `json:"name"` }
type createExpenseRequest struct { Description string `json:"description"`; Amount int64 `json:"amount"`; PaidBy string `json:"paidBy"`; Participants []string `json:"participants"` }
type Balance struct { MemberID string `json:"memberId"`; Balance int64 `json:"balance"` }
type Transfer struct { From string `json:"from"`; To string `json:"to"`; Amount int64 `json:"amount"` }

func main() {
	s := newStore(); mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status": "ok"}) })
	mux.HandleFunc("/api/groups", s.groupsHandler); mux.HandleFunc("/api/groups/", s.groupHandler)
	log.Printf("splitbill API listening on :8080"); log.Fatal(http.ListenAndServe(":8080", mux))
}

func (s *Store) groupsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/groups" { http.NotFound(w, r); return }
	if r.Method == http.MethodGet { s.RLock(); defer s.RUnlock(); result := make([]*Group, 0, len(s.groups)); for _, g := range s.groups { result = append(result, cloneGroup(g)) }; writeJSON(w, http.StatusOK, result); return }
	if r.Method != http.MethodPost { methodNotAllowed(w); return }
	var req createGroupRequest; if !decodeJSON(w, r, &req) { return }; if strings.TrimSpace(req.Name) == "" { writeError(w, http.StatusBadRequest, "name is required"); return }
	s.Lock(); defer s.Unlock(); g := &Group{ID: s.id("grp"), Name: strings.TrimSpace(req.Name), Members: []Member{}, Expenses: []Expense{}}; s.groups[g.ID] = g; writeJSON(w, http.StatusCreated, cloneGroup(g))
}

func (s *Store) groupHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/"); if len(parts) < 3 || parts[0] != "api" || parts[1] != "groups" { http.NotFound(w, r); return }
	groupID := parts[2]; s.RLock(); _, ok := s.groups[groupID]; s.RUnlock(); if !ok { writeError(w, http.StatusNotFound, "group not found"); return }
	if len(parts) == 3 { if r.Method != http.MethodGet { methodNotAllowed(w); return }; s.RLock(); g := cloneGroup(s.groups[groupID]); s.RUnlock(); writeJSON(w, http.StatusOK, g); return }
	if len(parts) != 4 { http.NotFound(w, r); return }
	switch parts[3] { case "members": s.membersHandler(w, r, groupID); case "expenses": s.expensesHandler(w, r, groupID); case "balances": s.summaryHandler(w, r, groupID, false); case "settlements": s.summaryHandler(w, r, groupID, true); default: http.NotFound(w, r) }
}

func (s *Store) membersHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method == http.MethodGet { s.RLock(); result := append([]Member(nil), s.groups[id].Members...); s.RUnlock(); writeJSON(w, http.StatusOK, result); return }
	if r.Method != http.MethodPost { methodNotAllowed(w); return }; var req createMemberRequest; if !decodeJSON(w, r, &req) { return }; name := strings.TrimSpace(req.Name); if name == "" { writeError(w, http.StatusBadRequest, "name is required"); return }
	s.Lock(); defer s.Unlock(); for _, m := range s.groups[id].Members { if m.Name == name { writeError(w, http.StatusConflict, "member name already exists"); return } }; m := Member{ID: s.id("mem"), Name: name}; s.groups[id].Members = append(s.groups[id].Members, m); writeJSON(w, http.StatusCreated, m)
}

func (s *Store) expensesHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method == http.MethodGet { s.RLock(); result := append([]Expense(nil), s.groups[id].Expenses...); s.RUnlock(); writeJSON(w, http.StatusOK, result); return }
	if r.Method != http.MethodPost { methodNotAllowed(w); return }; var req createExpenseRequest; if !decodeJSON(w, r, &req) { return }
	if req.Amount <= 0 || req.PaidBy == "" || len(req.Participants) == 0 { writeError(w, http.StatusBadRequest, "amount, paidBy, and participants are required"); return }
	s.Lock(); defer s.Unlock(); if !hasMember(s.groups[id], req.PaidBy) { writeError(w, http.StatusBadRequest, "paidBy is not a member"); return }; seen := map[string]bool{}
	for _, memberID := range req.Participants { if seen[memberID] || !hasMember(s.groups[id], memberID) { writeError(w, http.StatusBadRequest, "participants must be unique group members"); return }; seen[memberID] = true }
	e := Expense{ID: s.id("exp"), Description: strings.TrimSpace(req.Description), Amount: req.Amount, PaidBy: req.PaidBy, Participants: append([]string(nil), req.Participants...), CreatedAt: time.Now().UTC().Format(time.RFC3339)}; s.groups[id].Expenses = append(s.groups[id].Expenses, e); writeJSON(w, http.StatusCreated, e)
}

func (s *Store) summaryHandler(w http.ResponseWriter, r *http.Request, id string, settlement bool) {
	if r.Method != http.MethodGet { methodNotAllowed(w); return }; s.RLock(); g := cloneGroup(s.groups[id]); s.RUnlock(); balances := calculateBalances(g); if !settlement { writeJSON(w, http.StatusOK, balances); return }; writeJSON(w, http.StatusOK, calculateSettlements(balances))
}
func hasMember(g *Group, id string) bool { for _, m := range g.Members { if m.ID == id { return true } }; return false }
func cloneGroup(g *Group) *Group {
	c := *g
	c.Members = append([]Member{}, g.Members...)
	c.Expenses = append([]Expense{}, g.Expenses...)
	return &c
}

func calculateBalances(g *Group) []Balance {
	paid, owed := map[string]int64{}, map[string]int64{}; for _, m := range g.Members { paid[m.ID] = 0; owed[m.ID] = 0 }
	for _, e := range g.Expenses { paid[e.PaidBy] += e.Amount; base, remainder := e.Amount/int64(len(e.Participants)), e.Amount%int64(len(e.Participants)); for i, id := range e.Participants { owed[id] += base; if int64(i) < remainder { owed[id]++ } } }
	result := make([]Balance, 0, len(g.Members)); for _, m := range g.Members { result = append(result, Balance{MemberID: m.ID, Balance: paid[m.ID] - owed[m.ID]}) }; return result
}
func calculateSettlements(balances []Balance) []Transfer {
	creditors, debtors := []Balance{}, []Balance{}; for _, b := range balances { if b.Balance > 0 { creditors = append(creditors, b) }; if b.Balance < 0 { debtors = append(debtors, Balance{MemberID: b.MemberID, Balance: -b.Balance}) } }; result := []Transfer{}
	for d, c := 0, 0; d < len(debtors) && c < len(creditors); { amount := int64(math.Min(float64(debtors[d].Balance), float64(creditors[c].Balance))); result = append(result, Transfer{From: debtors[d].MemberID, To: creditors[c].MemberID, Amount: amount}); debtors[d].Balance -= amount; creditors[c].Balance -= amount; if debtors[d].Balance == 0 { d++ }; if creditors[c].Balance == 0 { c++ } }; return result
}
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool { if json.NewDecoder(r.Body).Decode(target) != nil { writeError(w, http.StatusBadRequest, "invalid JSON"); return false }; return true }
func writeJSON(w http.ResponseWriter, status int, value any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(value) }
func writeError(w http.ResponseWriter, status int, message string) { writeJSON(w, status, map[string]string{"error": message}) }
func methodNotAllowed(w http.ResponseWriter) { writeError(w, http.StatusMethodNotAllowed, "method not allowed") }