import React from 'react';
import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { inject, observer } from 'mobx-react';
import FormContent from './FormContent';

@inject('project')
@observer
class AddProject extends React.Component {
  // submit form data
  submit = values => {
    const { project, history } = this.props;
    const data = { ...values };
    data.spec.services = _.map(values.spec.services, n => {
      const resources = n.split('/');
      return { type: resources[0], name: resources[1] };
    });
    project.createProject(data, () => {
      history.replace(`/project`);
    });
  };

  validate = values => {
    const errors = {};

    if (!values.metadata.name) {
      errors.metadata = { name: 'Required' };
    }
    return errors;
  };

  render() {
    const { history } = this.props;
    const initValue = {
      metadata: { name: '', description: '' },
      spec: {
        services: [],
        quota: {
          limits: {
            cpu: '',
            memory: '',
          },
          requests: {
            cpu: '',
            memory: '',
          },
        },
      },
    };
    return (
      <Formik
        initialValues={initValue}
        validate={this.validate}
        onSubmit={this.submit}
        render={props => <FormContent {...props} history={history} />}
      />
    );
  }
}

AddProject.propTypes = {
  handleSubmit: PropTypes.func,
  setFieldValue: PropTypes.func,
  project: PropTypes.object,
  history: PropTypes.object,
};

export default AddProject;
