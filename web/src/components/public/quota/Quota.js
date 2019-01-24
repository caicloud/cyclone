import React from 'react';
import { Radio, Form, Row, Col } from 'antd';
import { Field, withFormik } from 'formik';
import InputWithUnit from '@/components/public/inputWithUnit';
import MakeField from '@/components/public/makeField';
import PropTypes from 'prop-types';
import { resourceValidate } from '@/components/public/consts/validate';
const FormItem = Form.Item;
const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;

const _RadioGroup = MakeField(RadioGroup);

const allocationMap = [
  {
    name: 'basic',
    value: {
      'requests.cpu': '0.5',
      'limits.cpu': '1',
      'requests.memory': '1GiB',
      'limits.memory': '2GiB',
    },
  },
  {
    name: 'middle',
    value: {
      'requests.cpu': '1',
      'limits.cpu': '2',
      'requests.memory': '2 GiB',
      'limits.memory': '4 GiB',
    },
  },
  {
    name: 'high',
    value: {
      'requests.cpu': '2',
      'limits.cpu': '4',
      'requests.memory': '4GiB',
      'limits.memory': '8GiB',
    },
  },
];

class Quota extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    field: PropTypes.object,
    values: PropTypes.object,
    label: PropTypes.string,
    onChange: PropTypes.func,
    form: PropTypes.object,
  };
  handleType = e => {
    const { setFieldValue, onChange } = this.props;
    const value = e.target.value;
    setFieldValue('type', value);
    onChange('');
  };

  handleConfigSelect = e => {
    const { setFieldValue, onChange } = this.props;
    const value = e.target.value;
    setFieldValue('config', value);
    onChange(value);
  };

  handleInputChange = (name, value) => {
    const { setFieldValue } = this.props;
    setFieldValue(name, value);
  };

  handleBlur = () => {
    const {
      values,
      form: { errors },
      onChange,
      field: { value },
    } = this.props;
    const customValues = values.custom;
    if (
      _.get(customValues, 'limits.cpu') &&
      _.get(customValues, 'limits.memory') &&
      _.get(customValues, 'requests.memory') &&
      _.get(customValues, 'requests.cpu') &&
      _.isEmpty(_.get(errors, 'custom'))
    ) {
      onChange(customValues);
    } else if (value) {
      onChange('');
    }
  };

  // TODO(qme): realize custom and validate residual quota
  render() {
    const {
      values: { config, type },
      label,
    } = this.props;
    return (
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
              <RadioButton value="recommend">
                {intl.get('allocation.recommend')}
              </RadioButton>
              <RadioButton value="custom">
                {intl.get('allocation.custom')}
              </RadioButton>
            </Field>
          </div>
          <div className="allocation-content">
            {type === 'recommend' ? (
              <Field
                name="config"
                component={_RadioGroup}
                value={config}
                onChange={this.handleConfigSelect}
                formItemLayout={{ wrapperCol: { span: 20 } }}
              >
                {allocationMap.map(a => (
                  <RadioButton value={a.value} key={a.name}>
                    <div>
                      <div className="content">
                        <div>
                          <span className="key">
                            {intl.get('allocation.cpuRequest')}
                          </span>
                          <span>{a.value['requests.cpu']} Core</span>
                        </div>
                        <div>
                          <span className="key">
                            {intl.get('allocation.cpuLimit')}
                          </span>
                          <span>{a.value['limits.cpu']} Core</span>
                        </div>
                        <div>
                          <span className="key">
                            {intl.get('allocation.memoryRequest')}
                          </span>
                          <span>
                            {parseFloat(a.value['requests.memory'])} GiB
                          </span>
                        </div>
                        <div>
                          <span className="key">
                            {intl.get('allocation.memoryLimit')}
                          </span>
                          <span>
                            {parseFloat(a.value['limits.memory'])} GiB
                          </span>
                        </div>
                      </div>
                      <div className="footer">
                        {intl.get(`allocation.${a.name}`)}
                      </div>
                    </div>
                  </RadioButton>
                ))}
              </Field>
            ) : (
              <div className="custom">
                <Row gutter={16}>
                  <Col span={12}>
                    <Field
                      name="custom.requests.cpu"
                      render={props => {
                        const name = props.field.name;
                        const _props = {
                          error: _.get(props, `form.errors.${name}`),
                          touched: _.get(props, `form.touched.${name}`),
                          field: props.field,
                        };
                        return (
                          <InputWithUnit
                            label={intl.get('allocation.cpuRequest')}
                            addonAfter="Core"
                            defaultAddon=""
                            className="cpu"
                            {..._props}
                            onChange={this.handleInputChange}
                            onBlur={this.handleBlur}
                          />
                        );
                      }}
                    />
                  </Col>
                  <Col span={12}>
                    <Field
                      name="custom.requests.memory"
                      render={props => {
                        const name = props.field.name;
                        const _props = {
                          error: _.get(props, `form.errors.${name}`),
                          touched: _.get(props, `form.touched.${name}`),
                          field: props.field,
                        };
                        return (
                          <InputWithUnit
                            defaultAddon="MiB"
                            className="memory"
                            label={intl.get('allocation.memoryRequest')}
                            addonAfter={[
                              { name: 'MiB', value: 'MiB' },
                              { name: 'GiB', value: 'GiB' },
                              { name: 'TiB', value: 'TiB' },
                            ]}
                            {..._props}
                            onChange={this.handleInputChange}
                            onBlur={this.handleBlur}
                          />
                        );
                      }}
                    />
                  </Col>
                </Row>
                <Row gutter={16}>
                  <Col span={12}>
                    <Field
                      name="custom.limits.cpu"
                      render={props => {
                        const name = props.field.name;
                        const _props = {
                          error: _.get(props, `form.errors.${name}`),
                          touched: _.get(props, `form.touched.${name}`),
                          field: props.field,
                        };
                        return (
                          <InputWithUnit
                            label={intl.get('allocation.cpuLimit')}
                            addonAfter="Core"
                            defaultAddon=""
                            className="cpu"
                            {..._props}
                            onChange={this.handleInputChange}
                            onBlur={this.handleBlur}
                          />
                        );
                      }}
                    />
                  </Col>
                  <Col span={12}>
                    <Field
                      name="custom.limits.memory"
                      render={props => {
                        const name = props.field.name;
                        const _props = {
                          error: _.get(props, `form.errors.${name}`),
                          touched: _.get(props, `form.touched.${name}`),
                          field: props.field,
                        };
                        return (
                          <InputWithUnit
                            defaultAddon="MiB"
                            label={intl.get('allocation.memoryLimit')}
                            className="memory"
                            addonAfter={[
                              { name: 'MiB', value: 'MiB' },
                              { name: 'GiB', value: 'GiB' },
                              { name: 'TiB', value: 'TiB' },
                            ]}
                            {..._props}
                            onChange={this.handleInputChange}
                            onBlur={this.handleBlur}
                          />
                        );
                      }}
                    />
                  </Col>
                </Row>
              </div>
            )}
          </div>
        </div>
      </FormItem>
    );
  }
}

export default withFormik({
  mapPropsToValues: () => ({ type: 'recommend' }),
  validate: values => {
    const errors = {};
    if (values.type === 'recommend' && _.isEmpty(values.config)) {
      errors.config = '请选择配置';
    }
    if (values.type === 'custom') {
      errors.custom = resourceValidate(values.custom);
    }
    return errors;
  },
  displayName: 'allocation', // a unique identifier for this form
})(Quota);
