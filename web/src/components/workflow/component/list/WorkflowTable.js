import { Button, Input, Table } from 'antd';
import { FormatTime } from '@/lib/util';
import { inject, observer } from 'mobx-react';
import MenuAction from '@/components/workflow/component/MenuAction';
import PropTypes from 'prop-types';

const Search = Input.Search;
const Fragment = React.Fragment;

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
    matchPath: PropTypes.string,
  };

  addWorkFlow = () => {
    const { project, history } = this.props;
    history.push(`/workflow/add?project=${project}`);
  };

  render() {
    const { project, data, matchPath, history } = this.props;
    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'metadata.name',
        key: 'name',
      },
      {
        title: intl.get('workflow.creator'),
        dataIndex: 'metadata.annotations["cyclone.dev/owner"]',
        key: 'owner',
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
        align: 'right',
        render: value => (
          <MenuAction
            projectName={project}
            workflowName={value}
            history={history}
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
          onRow={row => {
            return {
              onClick: () => {
                this.props.history.push(
                  `${matchPath}/${row.metadata.name}?project=${project}`
                );
              },
            };
          }}
          columns={columns}
          dataSource={[...data]}
        />
      </Fragment>
    );
  }
}

export default WorkflowTable;
