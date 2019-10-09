import moment from 'moment';
import { Formik, Field } from 'formik';
import MakeField from '@/components/public/makeField';
import { defaultFormItemLayout } from '@/lib/const';
import { inject, observer } from 'mobx-react';
import { Modal, Radio, Input, Form } from 'antd';
import PropTypes from 'prop-types';

const InputField = MakeField(Input);

@inject('workflow')
@observer
class RunWorkflow extends React.Component {
  static propTypes = {
    workflow: PropTypes.shape({
      deleteWorkflow: PropTypes.func,
    }),
    projectName: PropTypes.string,
    workflowName: PropTypes.string,
    handleModalClose: PropTypes.func,
    visible: PropTypes.bool,
  };

  constructor(props) {
    super(props);
    const { visible } = props;
    this.state = {
      visible,
      versionMethod: 'auto',
    };
  }

  closeModal = () => {
    const { handleModalClose } = this.props;
    this.setState({ visible: false });
    handleModalClose && handleModalClose(false);
  };

  submitForm = value => {
    const {
      workflow: { runWorkflow },
      handleModalClose,
      projectName,
      workflowName,
    } = this.props;
    const { versionMethod } = this.state;
    if (versionMethod === 'auto') {
      const version = moment().format('YYYYMMDDhhmmss');
      runWorkflow(
        projectName,
        workflowName,
        {
          metadata: {
            name: version,
          },
        },
        // the parameters of list workflowrun callback
        {
          sort: true,
          ascending: false,
        }
      );
    } else {
      runWorkflow(
        projectName,
        workflowName,
        {
          metadata: {
            name: value.version,
          },
        },
        // the parameters of list workflowrun callback
        {
          sort: true,
          ascending: false,
        }
      );
    }

    this.setState({ visible: false });
    handleModalClose(false);
  };

  onVersionMethodChange = e => {
    this.setState({ versionMethod: e.target.value });
  };

  render() {
    const { visible } = this.props;
    const { versionMethod } = this.state;

    return (
      <Formik
        initialValues={{}}
        validate={values => {
          let errors = {};
          if (versionMethod === 'manual' && !_.get(values, 'version')) {
            errors = {
              version: 'version is required',
            };
          }

          return errors;
        }}
        onSubmit={this.submitForm}
        render={props => (
          <Modal
            title={intl.get('workflow.runWorkflow')}
            visible={visible}
            onCancel={this.closeModal}
            onOk={props.handleSubmit}
            width={600}
          >
            <Form.Item
              label={intl.get('workflowrun.version.method')}
              {...defaultFormItemLayout}
              labelCol={{ span: 6 }}
              wrapperCol={{ span: 18 }}
            >
              <Radio.Group
                defaultValue={versionMethod}
                value={versionMethod}
                onChange={this.onVersionMethodChange}
              >
                <Radio.Button value="auto">
                  {intl.get('workflowrun.version.auto')}
                </Radio.Button>
                <Radio.Button value="manual">
                  {intl.get('workflowrun.version.manual')}
                </Radio.Button>
              </Radio.Group>
            </Form.Item>
            {versionMethod === 'manual' && (
              <Field
                label={intl.get('version')}
                name="version"
                required
                hasFeedback
                component={InputField}
              />
            )}
          </Modal>
        )}
      />
    );
  }
}

export default RunWorkflow;
