import React from 'react';
import { Layout, Menu, Spin } from 'antd';
import { inject, observer, PropTypes as MobxPropTypes } from 'mobx-react';
import PropTypes from 'prop-types';
import qs from 'query-string';
import WorkflowTable from './WorkflowTable';

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
    const {
      history: { location },
    } = this.props;
    const query = qs.parse(location.search);
    this.props.project.listProjects(list => {
      const firstProject =
        query.project || _.get(list, 'items.[0].metadata.name');
      this.props.workflow.listWorklow(firstProject);
      this.props.history.replace(`/workflow?project=${firstProject}`);
    });
  }

  filterByProject = ({ key }) => {
    const {
      workflow: { listWorklow },
    } = this.props;
    listWorklow(key);
    this.props.history.replace(`/workflow?project=${key}`);
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
            defaultSelectedKeys={[query.project]}
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
          <WorkflowTable
            project={query.project}
            data={_workflowList}
            history={this.props.history}
          />
        </Content>
      </Layout>
    );
  }
}

export default List;
