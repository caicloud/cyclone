import EllipsisMenu from '@/components/public/ellipsisMenu';
import { inject, observer } from 'mobx-react';
import { Table, Modal, Tag, Spin, Input } from 'antd';
import { FormatTime, TimeDuration } from '@/lib/util';
import PropTypes from 'prop-types';

const confirm = Modal.confirm;
const Fragment = React.Fragment;
const Search = Input.Search;

@inject('workflow')
@observer
class WorkflowRuns extends React.Component {
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
    listWorkflowRuns(projectName, workflowName);
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
        delelteWorkflowRun(projectName, workflowName, name);
      },
    });
  };

  search = val => {
    const {
      workflow: { listWorkflowRuns },
      projectName,
      workflowName,
    } = this.props;
    listWorkflowRuns(projectName, workflowName, { filter: `name=${val}` });
  };

  render() {
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
          <Search
            placeholder="input search text"
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
