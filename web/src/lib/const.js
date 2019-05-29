export const defaultFormItemLayout = {
  labelCol: { span: 4 },
  wrapperCol: { span: 16 },
};

export const noLabelItemLayout = {
  labelCol: { span: 0 },
  wrapperCol: { span: 24 },
};

export const modalFormItemLayout = {
  labelCol: { span: 6 },
  wrapperCol: { span: 16 },
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

export const argumentsParamtersField = [
  { name: 'image', value: '' },
  { name: 'cmd', value: '' },
];

export const customStageField = {
  name: '',
  spec: {
    pod: {
      inputs: {
        resources: [],
      },
      spec: {
        containers: [
          {
            image: '',
            command: [],
            args: [],
            env: [],
          },
        ],
      },
      outputs: {
        resources: [],
      },
    },
  },
};
