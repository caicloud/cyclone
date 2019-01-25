import React from 'react';
import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import { Radio, Form, Input } from 'antd';
const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;
const InputField = MakeField(Input);
const FormItem = Form.Item;

const _RadioGroup = MakeField(RadioGroup);
const validateMap = {
  Token: <Field label="Token" name="Token" component={InputField} />,
  UserPwd: (
    <div>
      <Field label="用户名" name="UserName" component={InputField} />
      <Field label="密码" name="Pwd" component={InputField} />
    </div>
  ),
};
export default class ValidateSelect extends React.Component {
  state = {
    type: 'Token',
  };
  handleType = e => {
    this.setState({
      type: e.target.value,
    });
  };
  render() {
    return (
      <div>
        <FormItem
          label="验证方式"
          required
          {...{
            labelCol: { span: 4 },
            wrapperCol: { span: 14 },
          }}
        >
          <Field
            name="type"
            value={this.state.type}
            component={_RadioGroup}
            onChange={this.handleType}
          >
            <RadioButton value="Token">Token</RadioButton>
            <RadioButton value="UserPwd">用户名密码</RadioButton>
          </Field>
        </FormItem>
        {validateMap[this.state.type]}
      </div>
    );
  }
}
