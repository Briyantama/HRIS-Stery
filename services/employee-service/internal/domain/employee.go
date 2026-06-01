package domain

import (
	"fmt"
	"net/mail"
	"time"
)

// EmploymentStatus enum.
type EmploymentStatus string

const (
	StatusActive     EmploymentStatus = "ACTIVE"
	StatusInactive   EmploymentStatus = "INACTIVE"
	StatusOnLeave    EmploymentStatus = "ON_LEAVE"
	StatusTerminated EmploymentStatus = "TERMINATED"
)

// IsValid returns true if the status is a valid EmploymentStatus.
func (s EmploymentStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusOnLeave, StatusTerminated:
		return true
	default:
		return false
	}
}

// ContractType enum.
type ContractType string

const (
	ContractPermanent ContractType = "PERMANENT"
	ContractFixedTerm ContractType = "FIXED_TERM"
	ContractFreelance ContractType = "FREELANCE"
	ContractIntern    ContractType = "INTERN"
)

// IsValid returns true if the contract type is a valid ContractType.
func (c ContractType) IsValid() bool {
	switch c {
	case ContractPermanent, ContractFixedTerm, ContractFreelance, ContractIntern:
		return true
	default:
		return false
	}
}

// Gender enum.
type Gender string

const (
	GenderMale        Gender = "MALE"
	GenderFemale      Gender = "FEMALE"
	GenderUnspecified Gender = "UNSPECIFIED"
)

// IsValid returns true if the gender is a valid Gender.
func (g Gender) IsValid() bool {
	switch g {
	case GenderMale, GenderFemale, GenderUnspecified:
		return true
	default:
		return false
	}
}

// Employee is the aggregate root for employee records.
type Employee struct {
	id              EmployeeID
	tenantID        TenantID
	email           string
	phone           string
	fullName        string
	gender          Gender
	birthDate       *time.Time
	nationalID      string
	taxID           string
	departmentID    DepartmentID
	positionID      PositionID
	managerID       *EmployeeID
	status          EmploymentStatus
	contractType    ContractType
	joinDate        time.Time
	terminationDate *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

// NewEmployee creates a new employee record.
func NewEmployee(
	tenantID TenantID,
	email, fullName string,
	departmentID DepartmentID,
	positionID PositionID,
) (*Employee, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if fullName == "" {
		return nil, fmt.Errorf("full name is required")
	}
	if departmentID.IsZero() {
		return nil, fmt.Errorf("department is required")
	}
	if positionID.IsZero() {
		return nil, fmt.Errorf("position is required")
	}

	now := time.Now().UTC()
	return &Employee{
		id:           GenerateEmployeeID(),
		tenantID:     tenantID,
		email:        email,
		fullName:     fullName,
		departmentID: departmentID,
		positionID:   positionID,
		status:       StatusActive,
		contractType: ContractPermanent,
		joinDate:     now,
		gender:       GenderUnspecified,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// RehydrateEmployee reconstructs an employee from persisted data.
func RehydrateEmployee(
	id EmployeeID,
	tenantID TenantID,
	email, phone, fullName string,
	gender Gender,
	birthDate *time.Time,
	nationalID, taxID string,
	departmentID DepartmentID,
	positionID PositionID,
	managerID *EmployeeID,
	status EmploymentStatus,
	contractType ContractType,
	joinDate time.Time,
	terminationDate *time.Time,
	createdAt, updatedAt time.Time,
) (*Employee, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if fullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	return &Employee{
		id:              id,
		tenantID:        tenantID,
		email:           email,
		phone:           phone,
		fullName:        fullName,
		gender:          gender,
		birthDate:       birthDate,
		nationalID:      nationalID,
		taxID:           taxID,
		departmentID:    departmentID,
		positionID:      positionID,
		managerID:       managerID,
		status:          status,
		contractType:    contractType,
		joinDate:        joinDate,
		terminationDate: terminationDate,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}, nil
}

// ID returns the employee's unique identifier.
func (e *Employee) ID() EmployeeID {
	return e.id
}

// TenantID returns the tenant this employee belongs to.
func (e *Employee) TenantID() TenantID {
	return e.tenantID
}

// Email returns the employee's email address.
func (e *Employee) Email() string {
	return e.email
}

// Phone returns the employee's phone number.
func (e *Employee) Phone() string {
	return e.phone
}

// FullName returns the employee's full name.
func (e *Employee) FullName() string {
	return e.fullName
}

// Gender returns the employee's gender.
func (e *Employee) Gender() Gender {
	return e.gender
}

// BirthDate returns the employee's birth date.
func (e *Employee) BirthDate() *time.Time {
	return e.birthDate
}

// NationalID returns the employee's national ID.
func (e *Employee) NationalID() string {
	return e.nationalID
}

// TaxID returns the employee's tax ID.
func (e *Employee) TaxID() string {
	return e.taxID
}

// DepartmentID returns the employee's department ID.
func (e *Employee) DepartmentID() DepartmentID {
	return e.departmentID
}

// PositionID returns the employee's position ID.
func (e *Employee) PositionID() PositionID {
	return e.positionID
}

// ManagerID returns the employee's manager ID (if any).
func (e *Employee) ManagerID() *EmployeeID {
	return e.managerID
}

// Status returns the employee's employment status.
func (e *Employee) Status() EmploymentStatus {
	return e.status
}

// ContractType returns the employee's contract type.
func (e *Employee) ContractType() ContractType {
	return e.contractType
}

// JoinDate returns the employee's join date.
func (e *Employee) JoinDate() time.Time {
	return e.joinDate
}

// TerminationDate returns the employee's termination date (if terminated).
func (e *Employee) TerminationDate() *time.Time {
	return e.terminationDate
}

// CreatedAt returns the creation timestamp.
func (e *Employee) CreatedAt() time.Time {
	return e.createdAt
}

// UpdatedAt returns the last update timestamp.
func (e *Employee) UpdatedAt() time.Time {
	return e.updatedAt
}

// IsActive returns true if the employee is active.
func (e *Employee) IsActive() bool {
	return e.status == StatusActive
}

// IsTerminated returns true if the employee is terminated.
func (e *Employee) IsTerminated() bool {
	return e.status == StatusTerminated
}

// Terminate marks the employee as terminated.
func (e *Employee) Terminate(terminationDate time.Time) error {
	if e.IsTerminated() {
		return fmt.Errorf("employee is already terminated")
	}
	e.status = StatusTerminated
	e.terminationDate = &terminationDate
	e.updatedAt = time.Now().UTC()
	return nil
}

// UpdateDepartment changes the employee's department.
func (e *Employee) UpdateDepartment(departmentID DepartmentID) error {
	if e.IsTerminated() {
		return fmt.Errorf("cannot update terminated employee")
	}
	if departmentID.IsZero() {
		return fmt.Errorf("department is required")
	}
	e.departmentID = departmentID
	e.updatedAt = time.Now().UTC()
	return nil
}

// UpdateStatus transitions the employee to a new status.
func (e *Employee) UpdateStatus(newStatus EmploymentStatus) error {
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid status: %s", newStatus)
	}
	if e.IsTerminated() && newStatus != StatusTerminated {
		return fmt.Errorf("cannot change status of terminated employee")
	}

	// Validate allowed transitions
	switch newStatus {
	case StatusActive:
		if e.status == StatusTerminated {
			return fmt.Errorf("cannot resurrect terminated employee")
		}
	case StatusTerminated:
		// Terminal state, allowed from any non-terminated state
	case StatusOnLeave:
		if e.status == StatusTerminated {
			return fmt.Errorf("terminated employee cannot go on leave")
		}
	case StatusInactive:
		// Valid transition
	}

	e.status = newStatus
	e.updatedAt = time.Now().UTC()
	return nil
}

// SetManager assigns a manager to the employee.
func (e *Employee) SetManager(managerID *EmployeeID) error {
	if e.IsTerminated() {
		return fmt.Errorf("cannot assign manager to terminated employee")
	}
	e.managerID = managerID
	e.updatedAt = time.Now().UTC()
	return nil
}

// SetPhoneNumber updates the employee's phone number.
func (e *Employee) SetPhoneNumber(phone string) error {
	if e.IsTerminated() {
		return fmt.Errorf("cannot update terminated employee")
	}
	e.phone = phone
	e.updatedAt = time.Now().UTC()
	return nil
}

// validateEmail checks email format.
func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}
	return nil
}
