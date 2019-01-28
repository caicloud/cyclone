import React from 'react';
import { Radio, Form, Row, Col } from 'antd';
import { Field, withFormik } from 'formik';
import InputWithUnit from '@/components/public/inputWithUnit';
import MakeField from '@/components/public/makeField';
import PropTypes from 'prop-types';
import { resourceValidate } from '@/consts/validate';
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
      'requests.memory': '1Gi',
      'limits.memory': '2Gi',
    },
  },
  {
    name: 'middle',
    value: {
      'requests.cpu': '1',
      'limits.cpu': '2',
      'requests.memory': '2Gi',
      'limits.memory': '4Gi',
    },
  },
  {
    name: 'high',
    value: {
      'requests.cpu': '2',
      'limits.cpu': '4',
      'requests.memory': '4Gi',
      'limits.memory': '8Gi',
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
  constructor(props) {
    super(props);
    const { field } = props;
    this.getInitialValues(field.value);
  }

  getInitialValues = value => {
    const configItem = _.find(allocationMap, n => _.isEqual(n.value, value));
    let init = {};
    if (configItem > -1) {
      init.config = value;
    } else {
      init.custom = {
        limits: {
          cpu: _.get(value, 'limits.cpu'),
          memory: _.get(value, 'limits.memory'),
        },
        requests: {
          cpu: _.get(value, 'requests.cpu'),
          memory: _.get(value, 'requests.memory'),
        },
      };
    }
  };

  handleType = e => {
    const { setFieldValue, onChange } = this.props;
    const value = e.target.value;
    setFieldValue('type', value);
    onChange('');
  };

  handleConfigSelect = e => {
    const { setFieldValue, onChange } = this.props;
    const name = e.target.value;
    setFieldValue('config', name);
    const item = _.find(allocationMap, n => (n.name = name));
    onChange(item.value);
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
      onChange({
        'limits.cpu': _.get(customValues, 'limits.cpu'),
        'limits.memory': _.get(customValues, 'limits.memory'),
        'requests.memory': _.get(customValues, 'requests.memory'),
        'requests.cpu': _.get(customValues, 'requests.cpu'),
      });
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
                  <RadioButton value={a.name} key={a.name}>
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
                      render={props => (
                        <InputWithUnit
                          label={intl.get('allocation.cpuRequest')}
                          className="cpu"
                          type="cpu"
                          {...props}
                          onChange={this.handleInputChange}
                          onBlur={this.handleBlur}
                        />
                      )}
                    />
                  </Col>
                  <Col span={12}>
                    <Field
                      name="custom.requests.memory"
                      render={props => (
                        <InputWithUnit
                          className="memory"
                          label={intl.get('allocation.memoryRequest')}
                          type="memory"
                          {...props}
                          onChange={this.handleInputChange}
                          onBlur={this.handleBlur}
                        />
                      )}
                    />
                  </Col>
                </Row>
                <Row gutter={16}>
                  <Col span={12}>
                    <Field
                      name="custom.limits.cpu"
                      render={props => (
                        <InputWithUnit
                          label={intl.get('allocation.cpuLimit')}
                          type="cpu"
                          className="cpu"
                          {...props}
                          onChange={this.handleInputChange}
                          onBlur={this.handleBlur}
                        />
                      )}
                    />
                  </Col>
                  <Col span={12}>
                    <Field
                      name="custom.limits.memory"
                      render={props => (
                        <InputWithUnit
                          label={intl.get('allocation.memoryLimit')}
                          className="memory"
                          type="memory"
                          {...props}
                          onChange={this.handleInputChange}
                          onBlur={this.handleBlur}
                        />
                      )}
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
  mapPropsToValues: props => {
    const initValue = { type: 'recommend' };
    if (props.update) {
      const quotaValue = props.field.value;
      const configItem = _.find(allocationMap, n =>
        _.isEqual(n.value, quotaValue)
      );
      if (configItem) {
        initValue.config = configItem.name;
      } else {
        initValue.custom = {
          limits: {
            cpu: _.get(quotaValue, 'limits.cpu'),
            memory: _.get(quotaValue, 'limits.memory'),
          },
          requests: {
            cpu: _.get(quotaValue, 'requests.cpu'),
            memory: _.get(quotaValue, 'requests.memory'),
          },
        };
        initValue.type = 'custom';
      }
    }
    return initValue;
  },
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
