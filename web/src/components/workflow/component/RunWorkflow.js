import { Formik, Field } from 'formik';
import MakeField from '@/components/public/makeField';
import { inject, observer } from 'mobx-react';
import { Modal, Input } from 'antd';
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
    visible: PropTypes.boolean,
  };

  constructor(props) {
    super(props);
    const { visible } = props;
    this.state = {
      visible,
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
    runWorkflow(projectName, workflowName, value);
    this.setState({ visible: false });
    handleModalClose(false);
  };

  render() {
    const { visible } = this.props;
    return (
      <Formik
        initialValues={{}}
        validate={values => {
          let errors = {};
          if (!_.get(values, 'metadata.name')) {
            errors = {
              metadata: {
                name: 'name is required',
              },
            };
          }

          return errors;
        }}
        onSubmit={this.submitForm}
        render={props => (
          <Modal
            title="Run Workflow"
            visible={visible}
            onCancel={this.closeModal}
            onOk={props.handleSubmit}
          >
            <Field
              label={intl.get('name')}
              name="metadata.name"
              required
              hasFeedback
              component={InputField}
            />
          </Modal>
        )}
      />
    );
  }
}

export default RunWorkflow;
