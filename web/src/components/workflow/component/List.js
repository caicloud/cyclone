import React from 'react';
import { Table, Button, Input, Layout, Menu, Spin, Modal } from 'antd';
import { inject, observer, PropTypes as MobxPropTypes } from 'mobx-react';
import EllipsisMenu from '../../public/ellipsisMenu';
import PropTypes from 'prop-types';
import qs from 'query-string';

const confirm = Modal.confirm;
const Search = Input.Search;
const MenuItemGroup = Menu.ItemGroup;
const { Content, Sider } = Layout;

@inject('workflow', 'project')
@observer
class List extends React.Component {
  static propTypes = {
    workflow: PropTypes.shape({
      workflowList: MobxPropTypes.observableArray,
      listWorklow: PropTypes.func,
    }),
    history: PropTypes.object,
    match: PropTypes.object,
    project: PropTypes.shape({
      listProjects: PropTypes.func,
      projectList: MobxPropTypes.objectOrObservableObject,
    }),
  };

  componentDidMount() {
    this.props.project.listProjects(list => {
      const firstProject = _.get(list, 'items.[0].metadata.name');
      this.props.workflow.listWorklow(firstProject);
      this.props.history.replace(`/workflow?project=${firstProject}`);
    });
  }

  addWorkFlow = () => {
    const {
      history: { location },
    } = this.props;
    this.props.history.push(`/workflow/add${location.search}`);
  };

  filterByProject = ({ key }) => {
    this.props.history.replace(`/workflow?project=${key}`);
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

  render() {
    const {
      workflow: { workflowList },
      project: { projectList },
      history: { location },
    } = this.props;
    const query = qs.parse(location.search);

    if (!projectList) {
      return <Spin />;
    }

    const projectItems = _.get(projectList, 'items', []);
    const _workflowList = _.get(workflowList, `${query.project}.items`, []);
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
              this.deleteWorkflow(query.project, value);
            }}
          />
        ),
      },
    ];

    return (
      <Layout style={{ background: '#fff' }}>
        <Sider
          width={160}
          style={{ background: '#fff', borderRight: '1px solid #e8e8e8' }}
        >
          <Menu
            mode="inline"
            style={{ borderRight: 0 }}
            onSelect={this.filterByProject}
            defaultSelectedKeys={[_.get(projectItems, `[0].metadata.name`)]}
          >
            <MenuItemGroup key="g1" title={intl.get('sideNav.project')}>
              {projectItems.map(o => (
                <Menu.Item key={_.get(o, 'metadata.name')}>
                  {_.get(o, 'metadata.name')}
                </Menu.Item>
              ))}
            </MenuItemGroup>
          </Menu>
        </Sider>
        <Content style={{ width: '100%', paddingLeft: '24px' }}>
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
            dataSource={[..._workflowList]}
          />
        </Content>
      </Layout>
    );
  }
}

export default List;
