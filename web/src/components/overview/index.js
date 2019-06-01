import React from 'react';
import moment from 'moment';
import { Spin, Card, Statistic, Row, Col } from 'antd';
import { inject, observer } from 'mobx-react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts';
import PropTypes from 'prop-types';

@inject('dashboard')
@observer
class Overview extends React.Component {
  static propTypes = {
    dashboard: PropTypes.object,
  };

  componentDidMount() {
    this.props.dashboard.getStorageUsage();
    this.props.dashboard.getProjects();
    this.props.dashboard.getWorkflows();
    this.props.dashboard.getWorkflowRuns();
  }

  dayStatics(workflowRuns) {
    const result = _.reduce(
      workflowRuns.items,
      (result, wfr) => {
        const day = moment(_.get(wfr, 'metadata.creationTimestamp')).format(
          'YYYY/MM/DD'
        );
        if (!result[day]) {
          result[day] = {
            succeeded: 0,
            failed: 0,
            running: 0,
          };
        }

        const status = _.get(wfr, 'status.overall.phase');
        if (status === 'Succeeded') {
          result[day].succeeded += 1;
        } else if (status === 'Failed' || status === 'Cancelled') {
          result[day].failed += 1;
        } else {
          result[day].running += 1;
        }

        return result;
      },
      {}
    );
    return result;
  }

  recentDays() {
    const today = Math.floor(moment.now().valueOf() / 1000);
    const unit = moment.duration(1, 'days').asSeconds();
    return _.map(_.range(today - 6 * unit, today + unit, unit), v => {
      return moment.unix(v).format('YYYY/MM/DD');
    });
  }

  render() {
    const { dashboard } = this.props;
    if (dashboard.loading) {
      return <Spin />;
    }

    const wfrStatics = this.dayStatics(dashboard.workflowRuns);
    const days = this.recentDays();
    const data = _.map(days, day => {
      const v = wfrStatics[day] || {
        succeeded: 0,
        failed: 0,
        running: 0,
      };
      return {
        date: day,
        succeeded: v.succeeded,
        failed: v.failed,
        running: v.running,
      };
    });

    return (
      <div>
        <Row gutter={16}>
          <Col span={6}>
            <Card>
              <Statistic
                style={{ textAlign: 'center' }}
                title={intl.get('dashboard.projectCount')}
                value={`${_.get(dashboard.projects, 'metadata.total') || '--'}`}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                style={{ textAlign: 'center' }}
                title={intl.get('dashboard.workflowCount')}
                value={`${_.get(dashboard.workflows, 'metadata.total') ||
                  '--'}`}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                style={{ textAlign: 'center' }}
                title={intl.get('dashboard.workflowrunCount')}
                value={`${_.get(dashboard.workflowRuns, 'metadata.total') ||
                  '--'}`}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card>
              <Statistic
                style={{ textAlign: 'center' }}
                title={intl.get('dashboard.storageUsage')}
                value={`${dashboard.storageUsage.used || '--'}`}
                suffix={` / ${dashboard.storageUsage.total || '--'}`}
              />
            </Card>
          </Col>
        </Row>
        <Card
          style={{ marginTop: 24 }}
          title={intl.get('dashboard.workflowrunTrend')}
          bordered={true}
        >
          <BarChart
            width={720}
            height={240}
            data={data}
            margin={{
              top: 5,
              right: 30,
              left: 20,
              bottom: 5,
            }}
          >
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="date" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar
              dataKey="failed"
              stackId="status"
              name={intl.get('dashboard.executionFailed')}
              fill="#ff4d4f"
            />
            <Bar
              dataKey="running"
              stackId="status"
              name={intl.get('dashboard.executionRunning')}
              fill="#36cfc9"
            />
            <Bar
              dataKey="succeeded"
              stackId="status"
              name={intl.get('dashboard.executionSucceeded')}
              fill="#73d13d"
            />
          </BarChart>
        </Card>
      </div>
    );
  }
}

export default Overview;
