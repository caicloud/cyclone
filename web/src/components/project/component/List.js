import { Table, Button, Input } from 'antd';
import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
import MenuAction from './MenuAction';

const Search = Input.Search;

@inject('project')
@observer
class List extends React.Component {
  static propTypes = {
    match: PropTypes.object,
    history: PropTypes.object,
    project: PropTypes.object,
  };
  componentDidMount() {
    const { project } = this.props;
    project.listProjects();
  }
  saveFormRef = formRef => {
    this.formRef = formRef;
  };

  showModal = () => {
    const { match } = this.props;
    this.props.history.push(`${match.path}/add`);
  };

  search = val => {
    const {
      project: { listProjects },
    } = this.props;
    listProjects({ filter: `name=${val}` });
  };

  handleCreate = () => {
    const form = this.formRef.props.form;
    form.validateFields((err, values) => {
      if (err) {
        return;
      }
      form.resetFields();
    });
  };

  render() {
    const { match, project, history } = this.props;
    const columns = [
      {
        title: intl.get('name'),
        dataIndex: 'metadata.annotations["cyclone.dev/alias"]',
        key: 'name',
      },
      {
        title: intl.get('creationTime'),
        dataIndex: 'metadata.creationTimestamp',
        key: 'creationTime',
        render: value => FormatTime(value),
      },
      {
        title: intl.get('project.workflowCount'),
        dataIndex: 'workflowCount',
        key: 'workflowCount',
      },
      {
        title: intl.get('action'),
        dataIndex: 'metadata.name',
        key: 'action',
        align: 'right',
        render: value => <MenuAction name={value} history={history} />,
      },
    ];
    return (
      <div>
        <div className="head-bar">
          <Button type="primary" onClick={this.showModal}>
            {intl.get('operation.add')}
          </Button>
          <Search
            placeholder="input search text"
            onSearch={this.search}
            style={{ width: 200 }}
          />
        </div>
        <Table
          columns={columns}
          rowKey={record => record.metadata.name}
          onRow={record => {
            return {
              onClick: () => {
                this.props.history.push(
                  `${match.path}/${record.metadata.name}`
                );
              },
            };
          }}
          dataSource={_.get(project, 'projectList.items', [])}
        />
      </div>
    );
  }
}

export default List;
