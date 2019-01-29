import React from 'react';
import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import { Radio, Form, Input, Row, Col } from 'antd';
const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;
const InputField = MakeField(Input);
const FormItem = Form.Item;

const _RadioGroup = MakeField(RadioGroup);

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
    const validateMap = {
      Token: (
        <FormItem>
          <Field
            label="Token"
            name="spec.inline.scm.token"
            required
            component={InputField}
          />
          <Row>
            <Col offset={4} span={18}>
              <p className="token-tip">
                {intl.get('integration.form.pleaseClick')}
                <a
                  href="https://github.com/settings/tokens"
                  rel="noopener noreferrer"
                  target="_blank"
                >
                  [Access Token]
                </a>
                {intl.get('integration.form.tokentip')}
              </p>
            </Col>
          </Row>
        </FormItem>
      ),
      UserPwd: (
        <FormItem>
          <Field
            label={intl.get('integration.form.username')}
            name="spec.inline.scm.user"
            required
            component={InputField}
          />
          <Field
            label={intl.get('integration.form.pwd')}
            name="spec.inline.scm.password"
            required
            component={InputField}
          />
        </FormItem>
      ),
    };
    return (
      <div>
        <FormItem
          label={intl.get('integration.form.scm.verificationMode')}
          className="validate-select"
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
              {intl.get('integration.form.scm.usernamepwd')}
            </RadioButton>
          </Field>
        </FormItem>
        {validateMap[this.state.type]}
      </div>
    );
  }
}
