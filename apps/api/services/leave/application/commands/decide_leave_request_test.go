package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
)

const testTenant = "550e8400-e29b-41d4-a716-446655440000"

// seedRequest stores a 3-day PENDING request owned by the returned employee.
func seedRequest(t *testing.T, repo *mockLeaveRequestRepo) (*domain.LeaveRequest, domain.EmployeeID) {
	t.Helper()
	employee := domain.GenerateEmployeeID()
	start := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	req, err := domain.NewLeaveRequest(domain.MustNewTenantID(testTenant), employee, domain.GenerateLeaveTypeID(), start, start.AddDate(0, 0, 2), 3, "trip", employee.String())
	if err != nil {
		t.Fatal(err)
	}
	repo.requests[req.ID().String()] = req
	return req, employee
}

func TestApproveLeaveRequest(t *testing.T) {
	ctx := context.Background()
	repo := newMockLeaveRequestRepo()
	pub := newMockEventPublisher()
	h := NewApproveLeaveRequestHandler(repo, pub)
	req, employee := seedRequest(t, repo)
	manager := domain.GenerateEmployeeID()

	// A requester may not approve their own leave, even when privileged.
	_, err := h.Handle(ctx, ApproveLeaveRequestCommand{TenantID: testTenant, LeaveRequestID: req.ID().String(), ApprovedByID: employee.String(), ActorID: employee.String()})
	if !errors.Is(err, domain.ErrCannotDecideOwnRequest) {
		t.Fatalf("self-approval err = %v, want ErrCannotDecideOwnRequest", err)
	}
	if len(repo.transitions) != 0 || len(pub.published) != 0 || !req.IsPending() {
		t.Fatalf("self-approval must change nothing: %d transitions, %d events, status %s", len(repo.transitions), len(pub.published), req.Status())
	}

	// A different approver moves pending to used, guarded on the PENDING status.
	if _, err := h.Handle(ctx, ApproveLeaveRequestCommand{TenantID: testTenant, LeaveRequestID: req.ID().String(), ApprovedByID: manager.String(), ActorID: manager.String()}); err != nil {
		t.Fatalf("approve: %v", err)
	}
	want := transition{request: req, from: domain.StatusPending, delta: domain.BalanceDelta{Used: 3, Pending: -3}}
	if len(repo.transitions) != 1 || repo.transitions[0] != want {
		t.Fatalf("transitions = %+v, want [%+v]", repo.transitions, want)
	}
}

func TestApproveLeaveRequestLostRace(t *testing.T) {
	repo := newMockLeaveRequestRepo()
	pub := newMockEventPublisher()
	h := NewApproveLeaveRequestHandler(repo, pub)
	req, _ := seedRequest(t, repo)
	repo.transitionErr = domain.ErrInvalidStatusTransition // another decision won between read and write

	_, err := h.Handle(context.Background(), ApproveLeaveRequestCommand{TenantID: testTenant, LeaveRequestID: req.ID().String(), ApprovedByID: domain.GenerateEmployeeID().String(), ActorID: "x"})
	if !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Fatalf("err = %v, want ErrInvalidStatusTransition", err)
	}
	if len(pub.published) != 0 {
		t.Fatal("no event may be published when the status guard rejects the write")
	}
}

func TestRejectLeaveRequest(t *testing.T) {
	ctx := context.Background()
	repo := newMockLeaveRequestRepo()
	pub := newMockEventPublisher()
	h := NewRejectLeaveRequestHandler(repo, pub)
	req, employee := seedRequest(t, repo)
	manager := domain.GenerateEmployeeID()

	_, err := h.Handle(ctx, RejectLeaveRequestCommand{TenantID: testTenant, LeaveRequestID: req.ID().String(), Reason: "no", ActorID: employee.String()})
	if !errors.Is(err, domain.ErrCannotDecideOwnRequest) {
		t.Fatalf("self-rejection err = %v, want ErrCannotDecideOwnRequest", err)
	}
	if len(repo.transitions) != 0 || len(pub.published) != 0 {
		t.Fatal("self-rejection must change nothing")
	}

	if _, err := h.Handle(ctx, RejectLeaveRequestCommand{TenantID: testTenant, LeaveRequestID: req.ID().String(), Reason: "no", ActorID: manager.String()}); err != nil {
		t.Fatalf("reject: %v", err)
	}
	want := transition{request: req, from: domain.StatusPending, delta: domain.BalanceDelta{Pending: -3}}
	if len(repo.transitions) != 1 || repo.transitions[0] != want {
		t.Fatalf("transitions = %+v, want [%+v]", repo.transitions, want)
	}
}

func TestCancelLeaveRequestReleasesWhereHeld(t *testing.T) {
	for _, tc := range []struct {
		name     string
		approved bool
		from     domain.LeaveStatus
		delta    domain.BalanceDelta
	}{
		{"pending", false, domain.StatusPending, domain.BalanceDelta{Pending: -3}},
		{"approved", true, domain.StatusApproved, domain.BalanceDelta{Used: -3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newMockLeaveRequestRepo()
			h := NewCancelLeaveRequestHandler(repo, newMockEventPublisher())
			req, _ := seedRequest(t, repo)
			if tc.approved {
				if err := req.Approve(domain.GenerateEmployeeID()); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := h.Handle(context.Background(), CancelLeaveRequestCommand{TenantID: testTenant, LeaveRequestID: req.ID().String(), Reason: "plans", ActorID: "x"}); err != nil {
				t.Fatal(err)
			}
			want := transition{request: req, from: tc.from, delta: tc.delta}
			if len(repo.transitions) != 1 || repo.transitions[0] != want {
				t.Fatalf("transitions = %+v, want [%+v]", repo.transitions, want)
			}
		})
	}
}

func TestCreateLeaveRequestOverlap(t *testing.T) {
	repo := newMockLeaveRequestRepo()
	bal := newMockLeaveBalanceRepo()
	pub := newMockEventPublisher()
	val := newMockEmployeeValidator()
	tenant := domain.MustNewTenantID(testTenant)
	employee, leaveType := domain.GenerateEmployeeID(), domain.GenerateLeaveTypeID()
	b, _ := domain.NewLeaveBalance(tenant, employee, leaveType, 2026, 20)
	bal.balances[b.ID().String()] = b
	val.validEmployees[testTenant+":"+employee.String()] = true
	repo.createErr = domain.ErrLeaveOverlapExists

	start := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	_, err := NewCreateLeaveRequestHandler(repo, bal, pub, val).Handle(context.Background(), CreateLeaveRequestCommand{
		TenantID: testTenant, EmployeeID: employee.String(), LeaveTypeID: leaveType.String(),
		StartDate: start, EndDate: start.AddDate(0, 0, 2), DaysCount: 3, Reason: "trip", ActorID: employee.String(),
	})
	if !errors.Is(err, domain.ErrLeaveOverlapExists) {
		t.Fatalf("err = %v, want ErrLeaveOverlapExists", err)
	}
	if len(pub.published) != 0 {
		t.Fatal("no event may be published for a rejected overlapping request")
	}
}
