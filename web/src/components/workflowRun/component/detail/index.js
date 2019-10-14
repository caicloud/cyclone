import { Spin } from 'antd';
import { toJS } from 'mobx';
import { inject, observer } from 'mobx-react';
import { FormatTime, TimeDuration } from '@/lib/util';
import Detail from '@/components/public/detail';
const { DetailHead, DetailHeadItem, DetailContent } = Detail;

@inject('workflow')
@observer
class RunDetail extends React.Component {
  componentDidMount() {
    const {
      workflow: { getWorkflowRunLog, workflowRunDetail, getWorkflowRun },
      match: { params },
    } = this.props;
    if (!_.get(workflowRunDetail, _.get(params, 'workflowRun'))) {
      getWorkflowRun(params);
      // getWorkflowRunLog(params, { stage: 'qme-wf-stg0', container: 'i1' });
    }
  }
  render() {
    const {
      workflow: { workflowRunDetail },
      match: { params },
    } = this.props;

    const detail = toJS(_.get(workflowRunDetail, _.get(params, 'workflowRun')));
    if (!detail) {
      return <Spin />;
    }
    console.error(detail);
    return (
      <Detail>
        <DetailHead headName={_.get(detail, 'metadata.name')}>
          <DetailHeadItem
            name={intl.get('status.name')}
            value={FormatTime(_.get(detail, 'metadata.creationTimestamp'))}
          />
          <DetailHeadItem
            name={intl.get('duration')}
            value={
              TimeDuration(
                _.get(detail, 'status.overall.startTime'),
                _.get(detail, 'status.overall.lastTransitionTime')
              ) || '--'
            }
          />
        </DetailHead>
        <DetailContent />
      </Detail>
    );
    return <div>123</div>;
  }
}

export default RunDetail;
