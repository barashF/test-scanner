package mocks

import (
	context "context"
	reflect "reflect"

	uuid "github.com/google/uuid"
	model "github.com/internships-backend/test-backend-barashF/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MockroomRepository is a mock of roomRepository interface.
type MockroomRepository struct {
	ctrl     *gomock.Controller
	recorder *MockroomRepositoryMockRecorder
	isgomock struct{}
}

// MockroomRepositoryMockRecorder is the mock recorder for MockroomRepository.
type MockroomRepositoryMockRecorder struct {
	mock *MockroomRepository
}

// NewMockroomRepository creates a new mock instance.
func NewMockroomRepository(ctrl *gomock.Controller) *MockroomRepository {
	mock := &MockroomRepository{ctrl: ctrl}
	mock.recorder = &MockroomRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockroomRepository) EXPECT() *MockroomRepositoryMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockroomRepository) Create(arg0 context.Context, arg1 *model.Room) (uuid.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockroomRepositoryMockRecorder) Create(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockroomRepository)(nil).Create), arg0, arg1)
}

// GetAll mocks base method.
func (m *MockroomRepository) GetAll(arg0 context.Context) ([]*model.Room, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", arg0)
	ret0, _ := ret[0].([]*model.Room)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockroomRepositoryMockRecorder) GetAll(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockroomRepository)(nil).GetAll), arg0)
}

// GetByID mocks base method.
func (m *MockroomRepository) GetByID(arg0 context.Context, arg1 uuid.UUID) (*model.Room, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", arg0, arg1)
	ret0, _ := ret[0].(*model.Room)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockroomRepositoryMockRecorder) GetByID(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockroomRepository)(nil).GetByID), arg0, arg1)
}
