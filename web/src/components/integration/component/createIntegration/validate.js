export const validateForm = values => {
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
      const scmType = _.get(values, 'spec.scm.type');
      const scmValidateType = _.get(values, 'spec.scm.validateType');
      if (scmType === 'GitHub' || scmType === 'GitLab') {
        if (!values.spec.scm.server) {
          spec.scm.server = intl.get('integration.form.error.server');
          errors['spec'] = spec;
        }
        if (scmValidateType === 'Token') {
          if (!values.spec.scm.token) {
            spec.scm.token = intl.get('integration.form.error.token');
            errors['spec'] = spec;
          }
        } else {
          if (!values.spec.scm.user) {
            spec.scm.user = intl.get('integration.form.error.user');
            errors['spec'] = spec;
          }
          if (!values.spec.scm.password) {
            spec.scm.password = intl.get('integration.form.error.pwd');
            errors['spec'] = spec;
          }
        }
      }

      if (scmType === 'SVN') {
        if (!values.spec.scm.server) {
          spec.scm.server = intl.get('integration.form.error.server');
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
    }

    if (type === 'sonarQube') {
      if (!values.spec.sonarQube.server) {
        spec.sonarQube.server = intl.get('integration.form.error.server');
        errors['spec'] = spec;
      }
      if (!values.spec.sonarQube.token) {
        spec.sonarQube.token = intl.get('integration.form.error.token');
        errors['spec'] = spec;
      }
    }

    if (type === 'dockerRegistry') {
      if (!values.spec.dockerRegistry.server) {
        spec.dockerRegistry.server = intl.get('integration.form.error.server');
        errors['spec'] = spec;
      }
      if (!values.spec.dockerRegistry.user) {
        spec.dockerRegistry.user = intl.get('integration.form.error.user');
        errors['spec'] = spec;
      }
      if (!values.spec.dockerRegistry.password) {
        spec.dockerRegistry.password = intl.get('integration.form.error.pwd');
        errors['spec'] = spec;
      }
    }
  }
  return errors;
};
