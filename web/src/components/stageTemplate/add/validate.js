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
  if (!values.metadata.name) {
    errors.metadata = { name: intl.get('validate.required') };
    errors['spec'] = spec;
  }
  const args = _.get(values, 'spec.pod.spec.containers');
  _.forEach(args, (v, i) => {
    if (v.command === '') {
      spec.pod.spec.containers[i].command = intl.get('validate.required');
      errors['spec'] = spec;
    }
    if (v.image === '') {
      spec.pod.spec.containers[i].image = intl.get('validate.required');
      errors['spec'] = spec;
    }
  });
  return errors;
};
