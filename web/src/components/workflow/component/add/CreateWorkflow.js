import { Steps, Button, Form } from 'antd';
import Graph from './Graph';
import BasicInfo from './BasicInfo';
import styles from '../index.module.less';
import classNames from 'classnames/bind';
import PropTypes from 'prop-types';
import { tranformStage } from '@/lib/util';

const styleCls = classNames.bind(styles);
const Step = Steps.Step;

const steps = [
  {
    title: '基础信息', //intl.get('workflow.basicInfo'),
    content: <BasicInfo />,
  },
  {
    title: '任务', //intl.get('workflow.task'),
    content: <Graph />,
  },
];

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

  next() {
    const current = this.state.current + 1;
    this.setState({ current });
  }

  prev() {
    const current = this.state.current - 1;
    this.setState({ current });
  }

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
    const { handleSubmit } = this.props;
    console.log('value', this.props.values);
    return (
      <Form>
        <Steps current={current} size="small">
          {steps.map(item => (
            <Step key={item.title} title={item.title} />
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
            <Button type="primary" onClick={() => this.next()}>
              {intl.get('next')}
            </Button>
          )}
          {current === steps.length - 1 && (
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
