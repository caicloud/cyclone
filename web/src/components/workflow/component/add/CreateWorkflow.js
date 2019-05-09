import { Steps, Button, Form } from 'antd';

import Graph from './Graph';
import BasicInfo from './BasicInfo';
import styles from '../index.module.less';
import classNames from 'classnames/bind';
import PropTypes from 'prop-types';

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
  };

  constructor(props) {
    super(props);
    this.state = {
      current: 0,
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

  getStepContent = current => {
    const { setFieldValue, values, handleDepend } = this.props;
    switch (current) {
      case 0: {
        return <BasicInfo />;
      }
      case 1: {
        return (
          <Graph
            setFieldValue={setFieldValue}
            values={values}
            setStageDepned={handleDepend}
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
