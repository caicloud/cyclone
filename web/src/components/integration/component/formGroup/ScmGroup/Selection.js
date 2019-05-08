import { Radio, Form } from 'antd';
import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import PropTypes from 'prop-types';
const FormItem = Form.Item;
const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;

const _RadioGroup = MakeField(RadioGroup);

export default class Selection extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    setTouched: PropTypes.func,
    field: PropTypes.object,
    values: PropTypes.object,
    label: PropTypes.string,
    submitCount: PropTypes.number,
    onChange: PropTypes.func,
  };
  handleType = e => {
    const {
      setFieldValue,
      field: { name },
    } = this.props;
    const value = e.target.value;
    setFieldValue(name, value);
    if (value === 'GitLab') {
      setFieldValue('spec.scm.server', 'https://gitlab.com');
    } else if (value === 'SVN') {
      setFieldValue('spec.scm.server', '');
    } else if (value === 'GitHub') {
      setFieldValue('spec.scm.server', 'https://github.com');
    } else if (value === 'Bitbucket') {
      setFieldValue('spec.scm.server', 'https://bitbucket.com');
    }
  };

  render() {
    const { label } = this.props;
    return (
      <div>
        <FormItem
          label={label}
          required
          {...{
            labelCol: { span: 4 },
            wrapperCol: { span: 14 },
          }}
        >
          <div className="u-scm-sellection">
            <div className="scm-type">
              <Field
                name="spec.scm.type"
                component={_RadioGroup}
                onChange={this.handleType}
              >
                <RadioButton value="GitHub">GitHub</RadioButton>
                <RadioButton value="GitLab">GitLab</RadioButton>
                <RadioButton value="Bitbucket">Bitbucket</RadioButton>
                <RadioButton value="SVN">SVN</RadioButton>
              </Field>
            </div>
          </div>
        </FormItem>
      </div>
    );
  }
}
