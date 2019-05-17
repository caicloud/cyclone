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
  const args = _.get(values, 'spec.pod.inputs.arguments');
  _.forEach(args, (v, i) => {
    if (v.value === '') {
      spec.pod.inputs.arguments[i].value = intl.get('validate.required');
      errors['spec'] = spec;
    }
  });
  return errors;
};
