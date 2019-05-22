import { Tabs } from 'antd';
import { inject, observer } from 'mobx-react';
import Detail from '@/components/public/detail';
import PropTypes from 'prop-types';
import { FormatTime } from '@/lib/util';
import { getQuery } from '@/lib/util';
import StageDepend from './StageDepend';

const { DetailHead, DetailHeadItem, DetailContent } = Detail;
const TabPane = Tabs.TabPane;

@inject('workflow')
@observer
class WorkflowDetail extends React.Component {
  static propTypes = {
    match: PropTypes.object,
    project: PropTypes.object,
    history: PropTypes.object,
    workflow: PropTypes.object,
  };
  constructor(props) {
    super(props);
    const {
      workflow: { getWorkflow },
      match: { params },
      history: { location },
    } = this.props;
    const query = getQuery(location.search);

    getWorkflow(query.project, params.workflowName);
  }

  render() {
    const {
      match: {
        params: { workflowName },
      },
      workflow: { workflowDetail },
    } = this.props;
    const detail = _.get(workflowDetail, workflowName);

    return (
      <Detail>
        <DetailHead headName={_.get(detail, 'metadata.name')}>
          <DetailHeadItem
            name={intl.get('creationTime')}
            value={FormatTime(_.get(detail, 'metadata.creationTimestamp'))}
          />
          <DetailHeadItem
            name={intl.get('description')}
            value={_.get(detail, 'metadata.annotations.description')}
          />
        </DetailHead>
        <DetailContent>
          <Tabs defaultActiveKey="workflow" type="card">
            <TabPane tab={intl.get('workflow.basicInfo')} key="workflow">
              <StageDepend detail={detail} />
            </TabPane>
          </Tabs>
        </DetailContent>
      </Detail>
    );
  }
}

export default WorkflowDetail;
