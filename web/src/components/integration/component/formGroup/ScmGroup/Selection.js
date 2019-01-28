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
    values: PropTypes.object,
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
    if (value === 'GitLab') {
      setFieldValue('spec.inline.scm.server', 'https://gitlab.com');
    } else if (value === 'SVN') {
      setFieldValue('spec.inline.scm.server', '');
    } else if (value === 'GitHub') {
      setFieldValue('spec.inline.scm.server', 'https://github.com');
    }
  };

  render() {
    const {
      label,
      values: {
        spec: {
          inline: {
            scm: { type },
          },
        },
      },
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
                name="spec.inline.scm.type"
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
