import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import PropTypes from 'prop-types';
import { Radio, Form, Input } from 'antd';
const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;
const InputField = MakeField(Input);
const FormItem = Form.Item;

const _RadioGroup = MakeField(RadioGroup);

export default class ValidateSelect extends React.Component {
  static propTypes = {
    values: PropTypes.object,
    field: PropTypes.object,
    setFieldValue: PropTypes.func,
  };
  handleType = e => {
    const {
      setFieldValue,
      field: { name },
    } = this.props;
    const value = e.target.value;
    setFieldValue(name, value);
  };
  render() {
    const validateMap = {
      Token: (
        <FormItem>
          <Field
            label="Token"
            name="spec.cluster.credential.bearerToken"
            required
            component={InputField}
          />
        </FormItem>
      ),
      UserPwd: (
        <FormItem>
          <Field
            label={intl.get('integration.form.username')}
            name="spec.cluster.credential.user"
            required
            component={InputField}
          />
          <Field
            label={intl.get('integration.form.pwd')}
            name="spec.cluster.credential.password"
            type="password"
            required
            component={InputField}
          />
        </FormItem>
      ),
    };
    const {
      values: {
        spec: {
          cluster: {
            credential: { validateType },
          },
        },
      },
    } = this.props;
    return (
      <div>
        <FormItem
          label={intl.get('integration.form.cluster.verificationMode')}
          className="validate-select"
          required
          {...{
            labelCol: { span: 4 },
            wrapperCol: { span: 14 },
          }}
        >
          <Field
            name="spec.cluster.credential.validateType"
            component={_RadioGroup}
            onChange={this.handleType}
          >
            <RadioButton value="Token">Token</RadioButton>
            <RadioButton value="UserPwd">
              {intl.get('integration.form.cluster.usernamepwd')}
            </RadioButton>
          </Field>
        </FormItem>
        {validateMap[validateType]}
      </div>
    );
  }
}
