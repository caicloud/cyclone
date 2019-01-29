import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { inject, observer } from 'mobx-react';
import integration from '../../../../store/integration';
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
    payload: PropTypes.object,
    history: PropTypes.object,
    setFieldValue: PropTypes.func,
    sourceType: PropTypes.string,
    initialFormData: PropTypes.object,
  };

  handleSelectChange = val => {
    const { setFieldValue } = this.props;
    setFieldValue('sourceType', val);
  };

  handleCancle = () => {
    const { history } = this.props;
    history.push('/integration');
  };

  initFormValue = () => {
    return {
      metadata: { name: '', description: '', creationTime: '' },
      sourceType: '',
      spec: {
        dockerRegistry: {
          password: '',
          server: '',
          user: '',
        },
        general: [
          {
            name: '',
            value: '',
          },
        ],
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
    if (!values.metadata.name) {
      errors.metadata = { name: intl.get('integration.form.error.name') };
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
    const dsubmitData = generateData(values, sourceType);
    integration.createIntegration(dsubmitData, () => {
      this.props.history.replace(`/integration`);
    });
  };

  render() {
    return (
      <Formik
        initialValues={this.initFormValue()}
        validate={this.validateForm}
        onSubmit={this.submit}
        render={props => (
          <FormContent {...props} handleCancle={this.handleCancle} />
        )}
      />
    );
  }
}
