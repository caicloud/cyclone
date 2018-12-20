import React from 'react';
import { Tabs } from 'antd';
import { inject, observer, PropTypes } from 'mobx-react';
import Detail from '@/public/detail/index';
const { DetailHead, DetailHeadItem, DetailContent } = Detail;

const TabPane = Tabs.TabPane;

@inject('workflow')
@observer
class ProjectDetail extends React.Component {
  static propTypes = {
    match: PropTypes.object,
  };
  render() {
    const {
      match: { params },
    } = this.props;
    return (
      <Detail>
        <DetailHead headName={params.projectId}>
          <DetailHeadItem name={intl.get('creationTime')} value="2018-09-08" />
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
