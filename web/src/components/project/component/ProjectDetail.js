import { Tabs, Spin } from 'antd';
import { inject, observer } from 'mobx-react';
import Detail from '@/components/public/detail';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
import MenuAction from './MenuAction';

const { DetailHead, DetailHeadItem, DetailContent, DetailAction } = Detail;
const TabPane = Tabs.TabPane;

@inject('project')
@observer
class ProjectDetail extends React.Component {
  static propTypes = {
    match: PropTypes.object,
    project: PropTypes.object,
    history: PropTypes.object,
  };
  constructor(props) {
    super(props);
    const {
      match: {
        params: { projectName },
      },
    } = this.props;
    this.props.project.getProject(projectName);
  }
  render() {
    const {
      project,
      match: {
        params: { projectName },
      },
      history,
    } = this.props;
    const loading = project.detailLoading;
    if (loading) {
      return <Spin />;
    }
    const detail = project.projectDetail;
    return (
      <Detail
        actions={
          <DetailAction>
            <MenuAction name={projectName} history={history} detail />
          </DetailAction>
        }
      >
        <DetailHead headName={_.get(detail, 'metadata.name')}>
          <DetailHeadItem
            name={intl.get('creationTime')}
            value={FormatTime(_.get(detail, 'metadata.creationTimestamp'))}
          />
        </DetailHead>
        <DetailContent>
          <Tabs defaultActiveKey="workflow" type="card">
            <TabPane tab={intl.get('sideNav.resource')} key="resource">
              Content of Tab Pane 1
            </TabPane>
            <TabPane tab="stage" key="stage">
              Content of Tab Pane 2
            </TabPane>
            <TabPane tab={intl.get('sideNav.workflow')} key="workflow">
              Content of Tab Pane 3
            </TabPane>
          </Tabs>
        </DetailContent>
      </Detail>
    );
  }
}

export default ProjectDetail;
