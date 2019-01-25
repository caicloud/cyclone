import React from 'react';
import { Radio, Form } from 'antd';
import { Field, withFormik } from 'formik';
import MakeField from '@/components/public/makeField';
import GitHub from './Forms/GitHub';
import PropTypes from 'prop-types';
const FormItem = Form.Item;
const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;

const _RadioGroup = MakeField(RadioGroup);

const ScmMap = {
  GitHub: <GitHub />,
  GitLab: <div>GitLab</div>,
  SVN: <div>SVN</div>,
};

class Selection extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    field: PropTypes.object,
    values: PropTypes.object,
    label: PropTypes.string,
    onChange: PropTypes.func,
  };
  handleType = e => {
    const { setFieldValue, onChange } = this.props;
    const value = e.target.value;
    setFieldValue('type', value);
    onChange(value);
  };

  // TODO(qme): realize custom and validate residual quota
  render() {
    const {
      values: { type },
      label,
    } = this.props;
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

export default withFormik({
  mapPropsToValues: () => ({ type: 'GitHub' }),
  validate: values => {
    const errors = {};
    return errors;
  },
  displayName: 'selection', // a unique identifier for this form
})(Selection);
