export const validateForm = values => {
  const errors = {};
  const spec = {
    pod: {
      inputs: {
        arguments: [{}, {}],
      },
      outputs: {},
    },
  };
  if (!values.metadata.alias) {
    errors.metadata = { alias: intl.get('validate.required') };
    errors['spec'] = spec;
  }
  if (values.spec.pod.inputs.resources.length <= 0) {
    spec.pod.inputs.resources = intl.get('validate.required');
    errors['spec'] = spec;
  }
  if (values.spec.pod.outputs.resources.length <= 0) {
    spec.pod.outputs.resources = intl.get('validate.required');
    errors['spec'] = spec;
  }
  const args = _.get(values, 'spec.pod.inputs.arguments');
  _.forEach(args, (v, i) => {
    if (v.value === '') {
      spec.pod.inputs.arguments[i].value = intl.get('validate.required');
      errors['spec'] = spec;
    }
  });
  return errors;
};
