import EllipsisMenu from '@/components/public/ellipsisMenu';
import { inject, observer } from 'mobx-react';
import { Table, Modal } from 'antd';
import { FormatTime } from '@/lib/util';
import PropTypes from 'prop-types';

const confirm = Modal.confirm;

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
        title: intl.get('creationTime'),
        dataIndex: 'metadata.creationTimestamp',
        key: 'creationTime',
        render: value => FormatTime(value),
      },
      {
        title: intl.get('action'),
        dataIndex: 'metadata.name',
        key: 'action',
        render: value => (
          <EllipsisMenu menuFunc={() => this.removeRunRecord(value)} />
        ),
      },
    ];
    return (
      <Table
        rowKey={row => row.metadata.name}
        columns={columns}
        dataSource={[...items]}
      />
    );
  }
}

export default WorkflowRuns;
