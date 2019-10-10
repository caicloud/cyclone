import EllipsisMenu from '@/components/public/ellipsisMenu';
import { inject, observer } from 'mobx-react';
import { Table, Modal, Tag, Spin, Input, Select } from 'antd';
import { FormatTime, TimeDuration } from '@/lib/util';
import PropTypes from 'prop-types';

const confirm = Modal.confirm;
const Fragment = React.Fragment;
const Search = Input.Search;
const { Option } = Select;

const listParams = {
  sort: true,
  ascending: false,
};

@inject('workflow')
@observer
class WorkflowRuns extends React.Component {
  constructor(props) {
    super(props);
    this.state = { status: 'all', query: '' };
  }

  static propTypes = {
    workflow: PropTypes.shape({
      delelteWorkflowRun: PropTypes.func,
      listWorkflowRuns: PropTypes.func,
    }),
    projectName: PropTypes.string,
    workflowName: PropTypes.string,
  };

  componentDidMount() {
    const {
      workflow: { listWorkflowRuns },
      projectName,
      workflowName,
    } = this.props;
    listWorkflowRuns(projectName, workflowName, listParams, data => {
      const items = _.get(data, 'items', []);
      const loop = this.shouldLoopRequest(items);
      if (loop && !this.workflowRunInterval) {
        this.pollRequest(5000);
      }
    });
  }

  pollRequest = time => {
    this.workflowRunInterval = window.setInterval(() => {
      const { query, status } = this.state;
      this.requestListWorkflowRuns({
        ...listParams,
        filter: `name=${query}${status === 'all' ? '' : `,status=${status}`}`,
        silent: true,
      });
    }, time);
  };

  shouldLoopRequest = list => {
    let loop = false;
    if (list.length > 0) {
      const statusList = list.map(v => {
        return _.get(v, 'status.overall.phase', '');
      });
      const nonTerminatedList = ['Running', 'Pending', 'Waiting', ''];
      nonTerminatedList.forEach(v => {
        if (_.includes(statusList, v)) {
          loop = true;
        }
      });
    }
    return loop;
  };

  componentDidUpdate(prevProps) {
    const {
      workflow: { workflowRuns },
      projectName,
      workflowName,
    } = this.props;
    const items = _.get(
      workflowRuns,
      [`${projectName}-${workflowName}`, 'items'],
      []
    );
    const loop = this.shouldLoopRequest(items);
    if (loop && !this.workflowRunInterval) {
      this.pollRequest(5000);
    }
    if (!loop && this.workflowRunInterval) {
      window.clearInterval(this.workflowRunInterval);
    }
  }

  componentWillUnmount() {
    if (this.workflowRunInterval) {
      window.clearInterval(this.workflowRunInterval);
    }
  }

  removeRunRecord = name => {
    const {
      workflow: { delelteWorkflowRun },
      projectName,
      workflowName,
    } = this.props;
    confirm({
      title: intl.get('confirmTip.remove', {
        resourceType: 'WorkflowRun',
        name,
      }),
      onOk() {
        delelteWorkflowRun(projectName, workflowName, name, listParams);
      },
    });
  };

  requestListWorkflowRuns = params => {
    const {
      workflow: { listWorkflowRuns },
      projectName,
      workflowName,
    } = this.props;
    listWorkflowRuns(projectName, workflowName, params);
  };

  doSearch = () => {
    const { query, status } = this.state;
    this.requestListWorkflowRuns({
      ...listParams,
      filter: `name=${query}${status === 'all' ? '' : `,status=${status}`}`,
    });
  };

  onStatusSelectChange = v => {
    this.setState({ status: v }, this.doSearch);
  };

  search = val => {
    this.setState({ query: val }, this.doSearch);
  };

  render() {
    const {
      workflow: { workflowRuns },
      projectName,
      workflowName,
    } = this.props;

    const { status } = this.state;

    const items = _.get(
      workflowRuns,
      [`${projectName}-${workflowName}`, 'items'],
      []
    );

    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'metadata.name',
        key: 'name',
      },
      {
        title: intl.get('status.name'),
        dataIndex: 'status.overall.phase',
        key: 'status',
        render: value => {
          if (value === 'Succeeded') {
            return (
              <Tag color="green">
                {intl.get(`status.${value.toLowerCase()}`)}
              </Tag>
            );
          } else if (value === 'Failed') {
            return (
              <Tag color="red">{intl.get(`status.${value.toLowerCase()}`)}</Tag>
            );
          } else if (value === 'Running') {
            return (
              <Tag color="cyan">
                {intl.get(`status.${value.toLowerCase()}`)}
              </Tag>
            );
          } else {
            return <Tag>{intl.get(`status.${value.toLowerCase()}`)}</Tag>;
          }
        },
      },
      {
        title: intl.get('creationTime'),
        dataIndex: 'metadata.creationTimestamp',
        key: 'creationTime',
        render: value => FormatTime(value),
      },
      {
        title: intl.get('duration'),
        dataIndex: 'metadata.creationTimestamp',
        key: 'duration',
        render: (value, item) => {
          const status = _.get(item, 'status.overall.phase');
          if (
            status === 'Succeeded' ||
            status === 'Failed' ||
            status === 'Cancelled'
          ) {
            const endTime = _.get(item, 'status.overall.lastTransitionTime');
            if (endTime) {
              return TimeDuration(value, endTime);
            } else {
              return '--';
            }
          }

          return <Spin size="small" />;
        },
      },
      {
        title: intl.get('action'),
        dataIndex: 'metadata.name',
        key: 'action',
        align: 'right',
        render: value => (
          <EllipsisMenu menuFunc={() => this.removeRunRecord(value)} />
        ),
      },
    ];
    return (
      <Fragment>
        <div className="head-bar right">
          <Select
            defaultValue="all"
            value={status}
            style={{ width: 120, marginRight: 16 }}
            onChange={this.onStatusSelectChange}
          >
            <Option value="all">{intl.get('status.all')}</Option>
            <Option value="succeeded">{intl.get('status.succeeded')}</Option>
            <Option value="running">{intl.get('status.running')}</Option>
            <Option value="failed">{intl.get('status.failed')}</Option>
          </Select>
          <Search
            placeholder="input record name query"
            onSearch={this.search}
            style={{ width: 200 }}
          />
        </div>
        <Table
          rowKey={row => row.metadata.name}
          columns={columns}
          dataSource={[...items]}
        />
      </Fragment>
    );
  }
}

export default WorkflowRuns;
