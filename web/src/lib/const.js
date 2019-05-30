export const defaultFormItemLayout = {
  labelCol: { span: 4 },
  wrapperCol: { span: 16 },
};

export const noLabelItemLayout = {
  labelCol: { span: 0 },
  wrapperCol: { span: 24 },
};

export const emptyLabelItemLayout = {
  wrapperCol: { span: 16, offset: 4 },
};

export const modalFormItemLayout = {
  labelCol: { span: 7 },
  wrapperCol: { span: 16 },
};

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
