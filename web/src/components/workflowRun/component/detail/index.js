import { Spin, Modal, Button } from 'antd';
import { toJS } from 'mobx';
import qs from 'query-string';
import { inject, observer } from 'mobx-react';
import {
  TimeDuration,
  tranformStage,
  formatWorkflowRunStage,
} from '@/lib/util';
import PropTypes from 'prop-types';
import Detail from '@/components/public/detail';
import Log from '@/components/public/log';
import Graph from '@/components/workflow/component/add/Graph';
const { DetailHead, DetailHeadItem, DetailContent } = Detail;

export const LINE_FEED = /(?:\r\n|\r|\n)/;

export const recordFinishStatus = ['Succeeded', 'Failed', 'Cancelled'];

@inject('workflow')
@observer
class RunDetail extends React.Component {
  static propTypes = {
    workflow: PropTypes.object,
    match: PropTypes.object,
  };
  constructor(props) {
    super(props);
    this.state = { showLog: false, activeStage: '' };
  }
  componentDidMount() {
    const {
      workflow: { workflowRunDetail, getWorkflowRun },
      match: { params },
    } = this.props;
    if (!_.get(workflowRunDetail, _.get(params, 'workflowRun'))) {
      getWorkflowRun(params);
    }
  }

  showStageLog = stageId => {
    this.setState({ showLog: true, activeStage: stageId });
  };

  getUrl = status => {
    const { activeStage } = this.state;
    const {
      match: { params },
    } = this.props;
    let requestQuery = {
      stage: activeStage,
      container: 'main',
    };
    if (recordFinishStatus.includes(status)) {
      const queryString = qs.stringify(requestQuery);

      return `/projects/${_.get(params, 'projectName')}/workflows/${_.get(
        params,
        'workflowName'
      )}/workflowruns/${_.get(params, 'workflowRun')}/logs?${queryString}`;
    } else {
      // Running 状态再去获取日志, Pending 和 Waiting 状态获取不到日志
      if (status === 'Running') {
        // TODO: streamlog
      }
    }
  };

  formatData = value => {
    const {
      workflow: { workflowRunDetail },
      match: { params },
    } = this.props;
    const status = _.get(
      workflowRunDetail,
      `${_.get(params, 'workflowRun')}.status.overall.phase`
    );
    if (recordFinishStatus.includes(status)) {
      const logs = _.compact(value.split(LINE_FEED));
      return logs;
    } else {
      const test = value ? _.compact(value.split(LINE_FEED)) : [];
      return test;
    }
  };
  render() {
    const {
      workflow: { workflowRunDetail },
      match: { params },
    } = this.props;
    const { showLog, activeStage } = this.state;
    const detail = toJS(_.get(workflowRunDetail, _.get(params, 'workflowRun')));
    if (!detail) {
      return <Spin />;
    }
    const status = _.get(detail, 'status.overall.phase');
    return (
      <Detail>
        <DetailHead headName={_.get(detail, 'metadata.name')}>
          <DetailHeadItem name={intl.get('status.name')} value={status} />
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
        <DetailContent>
          <Graph
            readOnly={true}
            className="graph-log"
            initialGraph={tranformStage(
              formatWorkflowRunStage(_.get(detail, 'status.stages')),
              _.get(detail, 'metadata.annotations.stagePosition')
            )}
            project={_.get(params, 'project')}
            workflowName={_.get(params, 'workflowName')}
            handleStageLog={this.showStageLog}
          />
        </DetailContent>
        <Modal
          title={`${activeStage} 的日志`}
          visible={showLog}
          width={800}
          bodyStyle={{ padding: '24px 0 24px 24px' }}
          footer={
            <Button
              onClick={() => {
                this.setState({ showLog: false });
              }}
            >
              关闭
            </Button>
          }
          onCancel={() => {
            this.setState({ showLog: false });
          }}
        >
          <Log parse={this.formatData} url={this.getUrl(status)} />
        </Modal>
      </Detail>
    );
  }
}

export default RunDetail;
