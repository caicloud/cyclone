import React from 'react';
import { Radio, Form } from 'antd';
import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import GitHub from './Forms/GitHub';
import GitLab from './Forms/GitLab';
import SVN from './Forms/SVN';
import PropTypes from 'prop-types';
const FormItem = Form.Item;
const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;

const _RadioGroup = MakeField(RadioGroup);

const ScmMap = {
  GitHub: <GitHub />,
  GitLab: <GitLab />,
  SVN: <SVN />,
};

export default class Selection extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    field: PropTypes.object,
    label: PropTypes.string,
    onChange: PropTypes.func,
  };
  handleType = e => {
    const {
      setFieldValue,
      field: { name },
    } = this.props;
    const value = e.target.value;
    setFieldValue(name, value);
  };

  // TODO(qme): realize custom and validate residual quota
  render() {
    const { field, label } = this.props;
    const type = field.value || 'GitHub';
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
          {/* TODO: split into sub-components */}
          <div className="u-resource-allocation">
            <div className="allocation-type">
              <Field
                name="type"
                value={type}
                component={_RadioGroup}
                onChange={this.handleType}
              >
                <RadioButton value="GitHub">GitHub</RadioButton>
                <RadioButton value="GitLab">GitLab</RadioButton>
                <RadioButton value="SVN">SVN</RadioButton>
              </Field>
            </div>
          </div>
        </FormItem>
        {ScmMap[type]}
      </div>
    );
  }
}
