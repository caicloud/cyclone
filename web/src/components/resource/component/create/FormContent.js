import { Field, FieldArray } from 'formik';
import PropTypes from 'prop-types';
import { Form, Input, Button, Row, Col, Checkbox, Switch } from 'antd';
import MakeField from '@/components/public/makeField';
import { defaultFormItemLayout } from '@/lib/const';
import SelectSourceType from './SelectSourceType';

const InputField = MakeField(Input);
const SelectField = MakeField(SelectSourceType);
const FormItem = Form.Item;

const FormContent = ({
  history,
  handleSubmit,
  update,
  setFieldValue,
  values,
}) => {
  const bindType = _.get(values, 'spec.bind.integrationType');
  return (
    <Form layout={'horizontal'} onSubmit={handleSubmit}>
      <Field
        label={intl.get('resource.type')}
        name="spec.type"
        component={InputField}
        disabled={update}
        required
      />
      <Field
        label={intl.get('resource.resolver')}
        name="spec.resolver"
        component={InputField}
        required
      />
      <FormItem
        label={intl.get('resource.operations')}
        {...defaultFormItemLayout}
      >
        <Field
          label={intl.get('resource.resolver')}
          name="spec.operations"
          render={obj => {
            return (
              <Checkbox.Group
                options={['pull', 'push']}
                defaultValue={obj.form.values.spec.operations}
                onChange={checkedValues => {
                  obj.form.values.spec.operations = checkedValues;
                }}
              />
            );
          }}
          required
        />
      </FormItem>
      <Field
        label={intl.get('resource.binding')}
        name="spec.bind.integrationType"
        handleSelectChange={val => {
          setFieldValue('spec.bind.integrationType', val);
        }}
        component={SelectField}
      />
      <FormItem
        label={intl.get('template.form.inputs.arguments')}
        {...defaultFormItemLayout}
      >
        <FieldArray
          name="spec.parameters"
          render={arrayHelpers => {
            return (
              <div>
                {_.get(values, 'spec.parameters', []).length > 0 && (
                  <Row gutter={16}>
                    <Col span={4}>{intl.get('name')}</Col>
                    {bindType && (
                      <Col span={8}>{intl.get('resource.binding')}</Col>
                    )}
                    <Col span={8}>{intl.get('description')}</Col>
                    <Col span={4}>{intl.get('required')}</Col>
                  </Row>
                )}
                {_.get(values, 'spec.parameters', []).map((a, index) => {
                  return (
                    <FormItem key={index}>
                      <Row gutter={16}>
                        <Col span={4}>
                          <Field
                            key={index}
                            name={`spec.parameters.${index}.name`}
                            component={InputField}
                            hasFeedback
                          />
                        </Col>
                        {bindType && (
                          <Col span={8}>
                            <Field
                              key={index}
                              name={`spec.parameters.${index}.binding`}
                              component={InputField}
                              hasFeedback
                            />
                          </Col>
                        )}
                        <Col span={8}>
                          <Field
                            key={index}
                            name={`spec.parameters.${index}.description`}
                            component={InputField}
                            hasFeedback
                          />
                        </Col>
                        <Col span={2}>
                          <Switch
                            onChange={val => {
                              setFieldValue(
                                `spec.parameters.${index}.required`,
                                val
                              );
                            }}
                            defaultChecked={!!a.required}
                          />
                        </Col>
                        <Col span={2}>
                          <Button
                            type="circle"
                            icon="delete"
                            onClick={() => arrayHelpers.remove(index)}
                          />
                        </Col>
                      </Row>
                    </FormItem>
                  );
                })}
                <FormItem>
                  <Button
                    icon="plus"
                    onClick={() => {
                      arrayHelpers.push({ name: '', description: '' });
                    }}
                  >
                    {intl.get('template.form.inputs.addArgs')}
                  </Button>
                </FormItem>
              </div>
            );
          }}
        />
      </FormItem>
      <Row>
        <Col span={4} />
        <Col span={12}>
          <div className="form-bottom-btn">
            <Button type="primary" htmlType="submit">
              {intl.get('operation.confirm')}
            </Button>
            <Button
              onClick={() => {
                history.push(`/resource`);
              }}
            >
              {intl.get('operation.cancel')}
            </Button>
          </div>
        </Col>
      </Row>
    </Form>
  );
};

FormContent.propTypes = {
  history: PropTypes.object,
  values: PropTypes.object,
  handleSubmit: PropTypes.func,
  setFieldValue: PropTypes.func,
  update: PropTypes.bool,
  resource: PropTypes.object,
  submitCount: PropTypes.number,
};

export default FormContent;
