export const defaultFormItemLayout = {
  labelCol: { span: 4 },
  wrapperCol: { span: 16 },
};

export const noLabelItemLayout = {
  labelCol: { span: 0 },
  wrapperCol: { span: 20 },
};

export const resourceParametersField = {
  SCM: [
    { name: 'SCM_TOKEN', value: '' },
    { name: 'SCM_URL', value: '' },
    { name: 'SCM_RESIVION', value: '' },
  ],
  DockerRegistry: [
    { name: 'IMAGE', value: '' },
    { name: 'USER', value: '' },
    { name: 'PASSWORD', value: '' },
  ],
};
