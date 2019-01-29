import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { inject, observer } from 'mobx-react';
import FormContent from './FormContent';

const generateData = data => {
  delete data['sourceType'];
  data['metadata']['creationTime'] = Date.now().toString();
  return data;
};

@inject('integration')
@observer
export default class IntegrationForm extends React.Component {
  static propTypes = {
    history: PropTypes.object,
    integration: PropTypes.object,
    initialFormData: PropTypes.object,
  };

  handleCancle = () => {
    const { history } = this.props;
    history.push('/integration');
  };

  initFormValue = () => {
    return {
      metadata: { alias: '', description: '', creationTime: '' },
      sourceType: '',
      spec: {
        dockerRegistry: {
          password: '',
          server: '',
          user: '',
        },
        scm: {
          password: '',
          server: 'https://github.com',
          token: '',
          type: 'GitHub',
          user: '',
        },
        sonarQube: {
          token: '',
          server: '',
        },
      },
    };
  };

  validateForm = values => {
    const errors = {};
    const spec = {
      scm: {},
      sonarQube: {},
      dockerRegistry: {},
    };
    if (!values.metadata.alias) {
      errors.metadata = { alias: intl.get('integration.form.error.alias') };
    }
    if (!values.sourceType) {
      errors.sourceType = intl.get('integration.form.error.sourceType');
    }

    if (!values.spec.scm.server) {
      spec.scm.server = intl.get('integration.form.error.server');
      errors['spec'] = spec;
    }

    if (!values.spec.scm.token) {
      spec.scm.token = intl.get('integration.form.error.token');
      errors['spec'] = spec;
    }

    if (!values.spec.scm.user) {
      spec.scm.user = intl.get('integration.form.error.user');
      errors['spec'] = spec;
    }
    if (!values.spec.scm.password) {
      spec.scm.password = intl.get('integration.form.error.pwd');
      errors['spec'] = spec;
    }
    return errors;
  };

  submit = values => {
    const { sourceType } = values;
    const { integration } = this.props;
    const dsubmitData = generateData(values, sourceType);
    integration.createIntegration(dsubmitData, () => {
      this.props.history.replace(`/integration`);
    });
  };

  render() {
    const initialValues = this.initFormValue();
    return (
      <Formik
        initialValues={initialValues}
        validate={this.validateForm}
        onSubmit={this.submit}
        render={props => (
          <FormContent {...props} handleCancle={this.handleCancle} />
        )}
      />
    );
  }
}
