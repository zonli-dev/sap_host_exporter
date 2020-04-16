package start_service

import (
	"strings"
	"testing"

	"github.com/SUSE/sap_host_exporter/internal/sapcontrol"
	"github.com/SUSE/sap_host_exporter/test/mock_sapcontrol"
	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockWebService := mock_sapcontrol.NewMockWebService(ctrl)
	currentSapInstance := sapcontrol.CurrentSapInstance{}

	_, err := NewCollector(mockWebService, currentSapInstance)

	assert.Nil(t, err)
}

func TestProcessesMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWebService := mock_sapcontrol.NewMockWebService(ctrl)
	mockWebService.EXPECT().GetProcessList().Return(&sapcontrol.GetProcessListResponse{
		Processes: []*sapcontrol.OSProcess{
			{
				Name:        "enserver",
				Description: "foobar",
				Dispstatus:  sapcontrol.STATECOLOR_GREEN,
				Textstatus:  "Running",
				Starttime:   "",
				Elapsedtime: "",
				Pid:         30787,
			},
			{
				Name:        "msg_server",
				Description: "foobar2",
				Dispstatus:  sapcontrol.STATECOLOR_YELLOW,
				Textstatus:  "Stopping",
				Starttime:   "",
				Elapsedtime: "",
				Pid:         30786,
			},
		},
	}, nil)
	mockWebService.EXPECT().GetSystemInstanceList().Return(&sapcontrol.GetSystemInstanceListResponse{}, nil)
	currentSapInstance := sapcontrol.CurrentSapInstance{
		SID:      "HA1",
		Number:   0,
		Name:     "ASCS",
		Hostname: "sapha1as",
	}

	expectedMetrics := `
	# HELP sap_start_service_processes The processes started by the SAP Start Service
	# TYPE sap_start_service_processes gauge
	sap_start_service_processes{instance_hostname="sapha1as",instance_name="ASCS",instance_number="0",name="enserver",pid="30787",sid="HA1",status="Running"} 2
	sap_start_service_processes{instance_hostname="sapha1as",instance_name="ASCS",instance_number="0",name="msg_server",pid="30786",sid="HA1",status="Stopping"} 3
	`

	var err error
	collector, err := NewCollector(mockWebService, currentSapInstance)
	assert.NoError(t, err)

	err = testutil.CollectAndCompare(collector, strings.NewReader(expectedMetrics), "sap_start_service_processes")
	assert.NoError(t, err)
}

func TestInstancesMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWebService := mock_sapcontrol.NewMockWebService(ctrl)
	mockWebService.EXPECT().GetSystemInstanceList().Return(&sapcontrol.GetSystemInstanceListResponse{
		Instances: []*sapcontrol.SAPInstance{
			{
				Hostname:      "sapha1as",
				InstanceNr:    0,
				HttpPort:      50013,
				HttpsPort:     50014,
				StartPriority: "1",
				Features:      "MESSAGESERVER|ENQUE",
				Dispstatus:    sapcontrol.STATECOLOR_GREEN,
			},
			{
				Hostname:      "sapha1er",
				InstanceNr:    10,
				HttpPort:      51013,
				HttpsPort:     51014,
				StartPriority: "0.5",
				Features:      "ENQREP",
				Dispstatus:    sapcontrol.STATECOLOR_GREEN,
			},
		},
	}, nil)
	mockWebService.EXPECT().GetProcessList().Return(&sapcontrol.GetProcessListResponse{}, nil)
	currentSapInstance := sapcontrol.CurrentSapInstance{
		SID:      "HA1",
		Number:   0,
		Name:     "ASCS",
		Hostname: "sapha1as",
	}

	expectedMetrics := `
	# HELP sap_start_service_instances All instances of the whole SAP system
	# TYPE sap_start_service_instances gauge
    sap_start_service_instances{features="MESSAGESERVER|ENQUE",instance_hostname="sapha1as",instance_name="ASCS",instance_number="0",sid="HA1",start_priority="1"} 2
	`

	var err error
	collector, err := NewCollector(mockWebService, currentSapInstance)
	assert.NoError(t, err)

	err = testutil.CollectAndCompare(collector, strings.NewReader(expectedMetrics), "sap_start_service_instances")
	assert.NoError(t, err)
}
