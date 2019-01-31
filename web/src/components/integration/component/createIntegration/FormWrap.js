import PropTypes from 'prop-types';
import { Formik } from 'formik';
import { inject, observer } from 'mobx-react';
import { toJS } from 'mobx';
import { Spin } from 'antd';
import DevTools from 'mobx-react-devtools';
import FormContent from './FormContent';

@inject('integration')
@observer
export default class IntegrationForm extends React.Component {
  constructor(props) {
    super(props);
    const {
      match: { params },
    } = props;
    this.update = !!_.get(params, 'integrationName');
    if (this.update) {
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

  generateData = data => {
    const metadata = {
      creationTime: Date.now().toString(),
      annotations: {
        'cyclone.io/description': _.get(data, 'metadata.description', ''),
        'cyclone.io/alias': _.get(data, 'metadata.alias', ''),
      },
    };
    const type = _.get(data, 'spec.type');
    const spec = _.pick(data.spec, [`${type}`, 'type']); // 只取type类型的表单
    return { metadata, spec };
  };

  initFormValue = () => {
    const integrationDetail = toJS(this.props.integration.integrationDetail);
    return this.mapRequestFormToInitForm(integrationDetail);
  };

  componentWillUnmount() {
    this.props.integration.resetIntegration();
  }

  generateSpecObj = data => {
    let defaultSpec = {
      scm: {
        server: 'https://github.com',
        type: 'GitHub',
      },
      dockerRegistry: {
        server: '',
        user: '',
        password: '',
      },
      sonarQube: {
        server: '',
        token: '',
      },
      type: '',
    };
    const type = _.get(data, 'spec.type');
    const spec = _.get(data, 'spec');
    const specData = _.pick(spec, [`${type}`, 'type']);
    return _.assign(defaultSpec, specData);
  };

  mapRequestFormToInitForm = data => {
    const alias = _.get(
      data,
      ['metadata', 'annotations', 'cyclone.io/alias'],
      ''
    );
    const description = _.get(
      data,
      ['metadata', 'annotations', 'cyclone.io/description'],
      ''
    );
    const creationTime = _.get(data, 'metadata.creationTimestamp', '');
    const spec = this.generateSpecObj(data);
    return {
      metadata: { alias, description, creationTime },
      spec,
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
    } else {
      const type = _.get(values, 'spec.type');
      if (type === 'scm') {
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
      }

      if (type === 'sonarQube') {
        if (!values.spec.sonarQube.server) {
          spec.sonarQube.server = intl.get('integration.form.error.server');
          errors['spec'] = spec;
        }
        if (!values.spec.sonarQube.token) {
          spec.sonarQube.token = intl.get('integration.form.error.server');
          errors['spec'] = spec;
        }
      }

      if (type === 'dockerRegistry') {
        if (!values.spec.dockerRegistry.server) {
          spec.dockerRegistry.server = intl.get(
            'integration.form.error.server'
          );
          errors['spec'] = spec;
        }
        if (!values.spec.dockerRegistry.user) {
          spec.dockerRegistry.user = intl.get('integration.form.error.server');
          errors['spec'] = spec;
        }
        if (!values.spec.dockerRegistry.password) {
          spec.dockerRegistry.password = intl.get(
            'integration.form.error.server'
          );
          errors['spec'] = spec;
        }
      }
    }
    return errors;
  };

  submit = values => {
    const { integration } = this.props;
    const submitData = this.generateData(values);
    if (this.update) {
      const {
        match: { params },
      } = this.props;
      integration.updateIntegration(submitData, params.integrationName, () => {
        this.props.history.replace(`/integration`);
      });
    } else {
      integration.createIntegration(submitData, () => {
        this.props.history.replace(`/integration`);
      });
    }
  };

  render() {
    if (this.props.integration.detailLoading) {
      return <Spin />;
    }
    const initialValues = this.initFormValue();
    return (
      <div>
        <Formik
          enableReinitialize={true}
          initialValues={initialValues}
          validate={this.validateForm}
          onSubmit={this.submit}
          render={props => (
            <FormContent {...props} handleCancle={this.handleCancle} />
          )}
        />
        <DevTools />
      </div>
    );
  }
}
