import { Steps, Button, Form } from 'antd';
import Graph from './Graph';
import BasicInfo from './BasicInfo';
import styles from '../index.module.less';
import classNames from 'classnames/bind';
import PropTypes from 'prop-types';
import { tranformStage } from '@/lib/util';
import { inject, observer } from 'mobx-react';

const styleCls = classNames.bind(styles);
const Step = Steps.Step;

@inject('workflow')
@observer
class App extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    values: PropTypes.object,
    handleDepend: PropTypes.func,
    handleSubmit: PropTypes.func,
    submitting: PropTypes.bool,
    setSubmitting: PropTypes.func,
    saveStagePostition: PropTypes.func,
    workFlowInfo: PropTypes.object,
    workflow: PropTypes.object,
    workflowName: PropTypes.string,
    project: PropTypes.string,
  };

  constructor(props) {
    super(props);
    const { workFlowInfo } = props;
    this.state = {
      current: 0,
      graph: _.isEmpty(workFlowInfo)
        ? {}
        : tranformStage(
            _.get(workFlowInfo, 'spec.stages'),
            _.get(workFlowInfo, 'metadata.annotations.stagePosition')
          ),
    };
  }

  componentDidUpdate(prevProps) {
    const { submitting, setSubmitting } = this.props;
    if (!submitting && submitting !== prevProps.submitting) {
      setSubmitting(submitting);
    }
  }

  next = update => {
    const {
      workflow: { updateWorkflow, workflowDetail },
      values,
      workflowName,
      project,
    } = this.props;
    const current = this.state.current + 1;
    this.setState({ current });
    if (update) {
      const detail = _.get(workflowDetail, workflowName);
      const prevDes = _.get(detail, 'metadata.annotations.description');
      const des = _.get(values, 'metadata.annotations.description');
      if (prevDes !== des) {
        const workflowData = {
          metadata: { name: workflowName, annotations: { description: des } },
          ..._.pick(detail, 'spec.stages'),
        };
        updateWorkflow(project, workflowName, workflowData);
      }
    }
  };

  prev = update => {
    const current = this.state.current - 1;
    this.setState({ current });
  };

  saveGraph = graphData => {
    this.setState({ graph: graphData });
  };

  getStepContent = current => {
    const {
      setFieldValue,
      values,
      handleDepend,
      saveStagePostition,
      workFlowInfo,
      project,
      workflowName,
    } = this.props;
    const { graph } = this.state;
    switch (current) {
      case 0: {
        return <BasicInfo />;
      }
      case 1: {
        return (
          <Graph
            setFieldValue={setFieldValue}
            values={values}
            initialGraph={graph}
            update={!_.isEmpty(workFlowInfo)}
            project={project}
            workflowName={workflowName}
            setStageDepned={handleDepend}
            updateStagePosition={saveStagePostition}
            saveGraphWhenUnmount={this.saveGraph}
          />
        );
      }
      default: {
        return null;
      }
    }
  };

  render() {
    const { current } = this.state;
    const { handleSubmit, workflowName } = this.props;
    const update = !!workflowName;
    const steps = [
      {
        title: `${intl.get('workflow.basicInfo')}`,
        content: <BasicInfo />,
      },
      {
        title: `${intl.get('workflow.task')}`,
        content: <Graph />,
      },
    ];

    return (
      <Form>
        <Steps current={current} size="small">
          {steps.map((item, i) => (
            <Step key={i} title={item.title} />
          ))}
        </Steps>
        <div
          className={styleCls('steps-content', {
            graph: current === 1,
          })}
        >
          {this.getStepContent(current)}
        </div>
        <div className="steps-action">
          {current < steps.length - 1 && (
            <Button type="primary" onClick={() => this.next(update)}>
              {intl.get('next')}
            </Button>
          )}
          {current === steps.length - 1 && !update && (
            <Button type="primary" onClick={handleSubmit}>
              {intl.get('confirm')}
            </Button>
          )}
          {current > 0 && (
            <Button style={{ marginLeft: 8 }} onClick={() => this.prev()}>
              {intl.get('prev')}
            </Button>
          )}
        </div>
      </Form>
    );
  }
}

export default App;
