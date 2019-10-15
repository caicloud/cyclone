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
      // http://localhost:8080/apis/v1alpha1/projects/{project}/workflows/{workflow}/workflowruns/{workflowrun}/logs
      getWorkflowRunLog(params, { stage: 'qme-wf-stg0', container: '' });
    }
  }
  render() {
    const {
      workflow: { workflowRunDetail, workflowRunLogs },
      match: { params },
    } = this.props;

    console.error('workflowRunLogs', JSON.stringify(workflowRunLogs));
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
