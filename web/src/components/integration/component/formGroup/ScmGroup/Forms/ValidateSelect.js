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
  Token: (
    <div>
      <Field
        label="Token"
        name="spec.inline.scm.token"
        required
        component={InputField}
      />
      <p className="elliot">{intl.get('integration.name')}</p>
    </div>
  ),
  UserPwd: (
    <div>
      <Field
        label={intl.get('integration.form.username')}
        name="spec.inline.scm.user"
        required
        component={InputField}
      />
      <Field
        label={intl.get('integration.form.password')}
        name="spec.inline.scm.password"
        required
        component={InputField}
      />
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
          label={intl.get('integration.form.scm.verificationMode')}
          required
          {...{
            labelCol: { span: 4 },
            wrapperCol: { span: 14 },
          }}
        >
          <Field
            name="validateType"
            value={this.state.type}
            component={_RadioGroup}
            onChange={this.handleType}
          >
            <RadioButton value="Token">Token</RadioButton>
            <RadioButton value="UserPwd">
              label={intl.get('integration.form.scm.usernamepwd')}
            </RadioButton>
          </Field>
        </FormItem>
        {validateMap[this.state.type]}
      </div>
    );
  }
}
