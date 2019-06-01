import { Tabs, Spin } from 'antd';
import { inject, observer } from 'mobx-react';
import Detail from '@/components/public/detail';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
import MenuAction from '../MenuAction';
import ResourceList from './resource';
import StageList from './stage';
import WorkflowTable from '@/components/workflow/component/list/WorkflowTable';

const { DetailHead, DetailHeadItem, DetailContent, DetailAction } = Detail;
const TabPane = Tabs.TabPane;

@inject('project', 'workflow')
@observer
class ProjectDetail extends React.Component {
  static propTypes = {
    match: PropTypes.object,
    project: PropTypes.object,
    history: PropTypes.object,
    workflow: PropTypes.object,
  };
  constructor(props) {
    super(props);
    const {
      match: {
        params: { projectName },
      },
    } = this.props;
    this.props.project.getProject(projectName);
    this.props.workflow.listWorklow(projectName);
  }
  render() {
    const {
      project,
      match: {
        params: { projectName },
      },
      workflow: { workflowList },
      history,
    } = this.props;
    const loading = project.detailLoading;
    if (loading) {
      return <Spin />;
    }
    const detail = project.projectDetail;
    const _workflowList = _.get(workflowList, `${projectName}.items`, []);
    return (
      <Detail
        actions={
          <DetailAction>
            <MenuAction name={projectName} history={history} detail />
          </DetailAction>
        }
      >
        <DetailHead
          headName={_.get(detail, 'metadata.annotations["cyclone.dev/alias"]')}
        >
          <DetailHeadItem
            name={intl.get('creationTime')}
            value={FormatTime(_.get(detail, 'metadata.creationTimestamp'))}
            history={history}
          />
        </DetailHead>
        <DetailContent>
          <Tabs defaultActiveKey="workflow" type="card">
            <TabPane tab={intl.get('sideNav.workflow')} key="workflow">
              <WorkflowTable
                project={projectName}
                data={_workflowList}
                history={this.props.history}
              />
            </TabPane>
            <TabPane tab={intl.get('resources')} key="resource">
              <ResourceList projectName={projectName} />
            </TabPane>
            <TabPane tab={intl.get('project.stage')} key="stage">
              <StageList projectName={projectName} />
            </TabPane>
          </Tabs>
        </DetailContent>
      </Detail>
    );
  }
}

export default ProjectDetail;
