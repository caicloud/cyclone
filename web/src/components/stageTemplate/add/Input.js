import PropTypes from 'prop-types';
import SectionCard from '@/components/public/sectionCard';
import ResourceType from './ResourceType';
import MakeField from '@/components/public/makeField';
import { Field, FieldArray } from 'formik';
import { Form, Input, Row, Col, Button } from 'antd';
import { defaultFormItemLayout } from '@/lib/const';

const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);
const FormItem = Form.Item;

const InputSection = props => {
  const { values, setFieldValue, errors, resourceTypes } = props;
  return (
    <SectionCard title={intl.get('input')}>
      <FormItem
        label={intl.get('template.form.inputs.arguments')}
        {...defaultFormItemLayout}
      >
        <FieldArray
          name="spec.pod.inputs.arguments"
          render={arrayHelpers => (
            <div>
              {_.get(values, 'spec.pod.inputs.arguments', []).length > 0 && (
                <Row gutter={16}>
                  <Col span={5}>{intl.get('name')}</Col>
                  <Col span={10}>{intl.get('value')}</Col>
                  <Col span={9}>{intl.get('description')}</Col>
                </Row>
              )}
              {_.get(values, 'spec.pod.inputs.arguments', []).map(
                (a, index) => {
                  const style =
                    a.name === 'cmd'
                      ? {
                          style: {
                            height: 150,
                          },
                        }
                      : {};
                  return (
                    <FormItem key={index}>
                      <Row gutter={16}>
                        <Col span={5}>
                          <Field
                            key={index}
                            name={`spec.pod.inputs.arguments.${index}.name`}
                            component={InputField}
                            hasFeedback
                          />
                        </Col>
                        <Col span={10}>
                          <Field
                            key={index}
                            name={`spec.pod.inputs.arguments.${index}.value`}
                            component={
                              a.name === 'cmd' ? TextareaField : InputField
                            }
                            {...style}
                            hasFeedback
                          />
                        </Col>
                        <Col span={7}>
                          <Field
                            key={index}
                            name={`spec.pod.inputs.arguments.${index}.description`}
                            component={InputField}
                            hasFeedback
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
                }
              )}
              <FormItem>
                <Button
                  icon="plus"
                  onClick={() => {
                    arrayHelpers.push({ name: '', value: '', description: '' });
                  }}
                >
                  {intl.get('template.form.inputs.addArgs')}
                </Button>
              </FormItem>
            </div>
          )}
        />
      </FormItem>
      <ResourceType
        path="spec.pod.inputs.resources"
        values={values}
        setFieldValue={setFieldValue}
        options={resourceTypes}
        errors={errors}
      />
    </SectionCard>
  );
};

InputSection.propTypes = {
  setFieldValue: PropTypes.func,
  values: PropTypes.object,
  resource: PropTypes.object,
  errors: PropTypes.object,
  resourceTypes: PropTypes.array,
};

export default InputSection;
