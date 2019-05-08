import EllipsisMenu from '@/components/public/ellipsisMenu';
import { Modal, Button, Input, Table } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';

const Search = Input.Search;
const Fragment = React.Fragment;
const confirm = Modal.confirm;

@inject('workflow')
@observer
class WorkflowTable extends React.Component {
  static propTypes = {
    workflow: PropTypes.shape({
      deleteWorkflow: PropTypes.func,
    }),
    project: PropTypes.string,
    data: PropTypes.array,
    history: PropTypes.object,
  };

  deleteWorkflow = (project, workflow) => {
    const {
      workflow: { deleteWorkflow },
    } = this.props;
    confirm({
      title: `Do you Want to delete workflow ${workflow} ?`,
      onOk() {
        deleteWorkflow(project, workflow);
      },
    });
  };

  addWorkFlow = () => {
    const { project, history } = this.props;
    history.push(`/workflow/add?project=${project}`);
  };

  render() {
    const { project, data } = this.props;
    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'metadata.name',
        key: 'name',
      },
      {
        title: intl.get('workflow.recentVersion'),
        dataIndex: 'recentVersion',
        key: 'recentVersion',
      },
      {
        title: intl.get('workflow.creator'),
        dataIndex: 'owner',
        key: 'owner',
      },
      {
        title: intl.get('creationTime'),
        dataIndex: 'creationTime',
        key: 'creationTime',
      },
      {
        title: intl.get('action'),
        dataIndex: 'metadata.name',
        key: 'action',
        render: value => (
          <EllipsisMenu
            menuFunc={() => {
              this.deleteWorkflow(project, value);
            }}
          />
        ),
      },
    ];
    return (
      <Fragment>
        <div className="head-bar">
          <Button type="primary" onClick={this.addWorkFlow}>
            {intl.get('operation.add')}
          </Button>
          <Search
            placeholder="input search text"
            onSearch={() => {}}
            style={{ width: 200 }}
          />
        </div>
        <Table
          rowKey={row => row.id}
          columns={columns}
          dataSource={[...data]}
        />
      </Fragment>
    );
  }
}

export default WorkflowTable;
