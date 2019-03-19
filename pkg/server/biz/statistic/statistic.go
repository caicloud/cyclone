package statistic

import (
	"fmt"
	"strconv"
	"time"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
)

// Stats counts every StatusPhase number of wfrs, and calculate success ratio of the workflowruns.
func Stats(list *v1alpha1.WorkflowRunList, startTime, endTime string) (*api.Statistic, error) {
	start, end, err := checkAndTransTimes(startTime, endTime)
	if err != nil {
		return nil, err
	}

	log.Infof("stats wfrs length: %d, start time: %s, end time: %s", len(list.Items), start.String(), end.String())
	// filter all wfr by : start < creationTimestamp < end
	wfrs := make([]v1alpha1.WorkflowRun, 0)
	for _, wfr := range list.Items {
		t := wfr.CreationTimestamp.Time
		if t.Before(end) && t.After(start) {
			wfrs = append(wfrs, wfr)
		}
	}

	statistics := &api.Statistic{
		Overview: api.StatsOverview{
			Total:        len(wfrs),
			SuccessRatio: "0.00%",
		},
		Details: []*api.StatsDetail{},
	}

	initStatsDetails(statistics, start, end)

	for _, wfr := range wfrs {
		for _, detail := range statistics.Details {
			if detail.Timestamp == formatTimeToDay(wfr.CreationTimestamp.Time) {
				// set details status
				detail.StatsPhase = statsStatus(detail.StatsPhase, wfr.Status.Overall.Phase)
			}

		}

		// set overview status
		statistics.Overview.StatsPhase = statsStatus(statistics.Overview.StatsPhase, wfr.Status.Overall.Phase)
	}

	if statistics.Overview.Total != 0 {
		statistics.Overview.SuccessRatio = fmt.Sprintf("%.2f%%",
			float64(statistics.Overview.Completed)/float64(statistics.Overview.Total)*100)
	}
	return statistics, nil
}

// checkAndTransTimes validates start, end times; and translates them to time.Time format.
func checkAndTransTimes(start, end string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	if start == "" || end == "" {
		err := fmt.Errorf(" `startTime` and `endTime` can not be empty")
		return startTime, endTime, err
	}

	startTime, endTime, err := transTimes(start, end)
	if err != nil {
		err := fmt.Errorf("`startTime` and `endTime` must be int positive integer")
		return startTime, endTime, err
	}

	if startTime.After(endTime) {
		err := fmt.Errorf("`startTime` must less or equal than `endTime`")
		return startTime, endTime, err
	}
	return startTime, endTime, nil
}

// transTimes trans startTime and endTime from string to time.Time.
func transTimes(start, end string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time

	startInt, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		return startTime, endTime, err
	}
	startTime = time.Unix(startInt, 0)

	endInt, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		return startTime, endTime, err
	}
	endTime = time.Unix(endInt, 0)

	return startTime, endTime, nil
}

func initStatsDetails(statistics *api.Statistic, start, end time.Time) {
	for ; !start.After(end); start = start.Add(24 * time.Hour) {
		detail := &api.StatsDetail{
			Timestamp: formatTimeToDay(start),
		}
		statistics.Details = append(statistics.Details, detail)
	}

	// if last day not equal end day, append end day.
	endDay := formatTimeToDay(end)
	length := len(statistics.Details)
	if length > 0 {
		if statistics.Details[length-1].Timestamp != endDay {
			detail := &api.StatsDetail{
				Timestamp: endDay,
			}
			statistics.Details = append(statistics.Details, detail)
		}
	}
}

func formatTimeToDay(t time.Time) int64 {
	timestamp := t.Unix()
	return timestamp - (timestamp % 86400)
}

func statsStatus(s api.StatsPhase, recordStatus v1alpha1.StatusPhase) api.StatsPhase {
	switch recordStatus {
	case v1alpha1.StatusCompleted:
		s.Completed++
	case v1alpha1.StatusError:
		s.Error++
	case v1alpha1.StatusCancelled:
		s.Cancelled++
	case v1alpha1.StatusRunning:
		s.Running++
	case v1alpha1.StatusPending:
		s.Pending++
	case v1alpha1.StatusWaiting:
		s.Waiting++
	default:
	}

	return s
}
