import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { inject, observer } from 'mobx-react';
import FormContent from './FormContent';

const generateData = data => {
  data.metadata = {
    creationTime: Date.now().toString(),
    annotations: {
      'cyclone.io/description': _.get(data, 'metadata.description', ''),
      'cyclone.io/alias': _.get(data, 'metadata.alias', ''),
    },
  };
  return data;
};

@inject('integration')
@observer
export default class IntegrationForm extends React.Component {
  constructor(props) {
    super(props);
    const {
      match: { params },
    } = props;
    const update = !!_.get(params, 'integrationName');
    this.state = {
      update,
    };
    if (update) {
      props.integration.getIntegration(params.integrationName);
    }
  }

  static propTypes = {
    history: PropTypes.object,
    match: PropTypes.object,
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
        type: '',
      },
    };
  };

  validateForm = values => {
    const errors = {};
    const spec = {
      scm: {},
      sonarQube: {},
      dockerRegistry: {},
      type: '',
    };
    if (!values.metadata.alias) {
      errors.metadata = { alias: intl.get('integration.form.error.alias') };
    }

    if (!values.spec.type) {
      spec.type = intl.get('integration.form.error.sourceType');
      errors['spec'] = spec;
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
    const { integration } = this.props;
    const submitData = generateData(values);
    integration.createIntegration(submitData, () => {
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
