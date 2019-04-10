import PropTypes from 'prop-types';
import { Formik } from 'formik';
import CreateWorkflow from './CreateWorkflow';

class AddWorkflow extends React.Component {
  // submit form data
  submit = () => {};

  validate = () => {
    const errors = {};
    return errors;
  };

  getInitialValues = () => {
    let defaultValue = {
      metadata: { alias: '', description: '' },
    };
    return defaultValue;
  };

  render() {
    const initValue = this.getInitialValues();
    return (
      <Formik
        initialValues={initValue}
        validate={this.validate}
        onSubmit={this.submit}
        render={props => <CreateWorkflow {...props} />}
      />
    );
  }
}

AddWorkflow.propTypes = {
  handleSubmit: PropTypes.func,
};

export default AddWorkflow;
