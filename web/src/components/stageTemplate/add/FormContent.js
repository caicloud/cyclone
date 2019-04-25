import PropTypes from 'prop-types';
import { Form, Input, Button, Col, Row } from 'antd';
import SectionCard from '@/components/public/sectionCard';
import { Field, FieldArray } from 'formik';
import { defaultFormItemLayout } from '@/lib/const';
import MakeField from '@/components/public/makeField';
import SelectSourceType from './SelectSourceType';

const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);
const SelectField = MakeField(SelectSourceType);
const FormItem = Form.Item;
const Fragment = React.Fragment;
const FormContent = props => {
  const { setFieldValue, handleCancle, handleSubmit, update, values } = props;
  return (
    <Form onSubmit={handleSubmit}>
      <Field
        label={intl.get('name')}
        name="metadata.alias"
        component={InputField}
        disabled={update}
        required
      />
      <Field
        label={intl.get('description')}
        name="metadata.description"
        component={TextareaField}
      />
      <SectionCard title={intl.get('input')}>
        <FormItem>
          <Field
            label={intl.get('type')}
            name="spec.pod.inputs.resources"
            required
            handleSelectChange={val => {
              setFieldValue('spec.pod.inputs.resources', [
                {
                  name: '',
                  type: val,
                  path: '',
                },
              ]);
            }}
            component={SelectField}
          />
        </FormItem>
      </SectionCard>
      <SectionCard title={intl.get('config')}>
        <FieldArray
          name="spec.pod.inputs.arguments"
          render={() => (
            <Fragment>
              {_.get(values, 'spec.pod.inputs.arguments', []).map(
                (field, index) => (
                  <Field
                    key={field.name}
                    label={intl.get(`template.form.config.${field.name}`)}
                    name={`spec.pod.inputs.arguments.${index}.value`}
                    component={InputField}
                    required
                  />
                )
              )}
            </Fragment>
          )}
        />
        <FieldArray
          name="spec.pod.spec.containers"
          render={() => (
            <div>
              {_.get(values, 'spec.pod.spec.containers', []).map((a, index) => (
                <Fragment key={index}>
                  <FormItem label={intl.get('env')} {...defaultFormItemLayout}>
                    <FieldArray
                      name={`spec.pod.spec.containers.${index}.env`}
                      render={arrayHelpers => (
                        <div>
                          {_.get(
                            values,
                            `spec.pod.spec.containers.${index}.env`,
                            []
                          ).length > 0 && (
                            <Row gutter={16}>
                              <Col span={11}>{intl.get('key')}</Col>
                              <Col span={11}>{intl.get('value')}</Col>
                            </Row>
                          )}
                          {_.get(
                            values,
                            `spec.pod.spec.containers.${index}.env`,
                            []
                          ).map((a, i) => (
                            <Row key={i} gutter={16}>
                              <Col span={11}>
                                <Field
                                  key={a.name}
                                  name={`spec.pod.spec.containers.${index}.env.${i}.name`}
                                  component={InputField}
                                  hasFeedback
                                />
                              </Col>
                              <Col span={11}>
                                <Field
                                  key={a.value}
                                  name={`spec.pod.spec.containers.${index}.env.${i}.value`}
                                  component={InputField}
                                  hasFeedback
                                />
                              </Col>
                              <Col span={2}>
                                <Button
                                  type="circle"
                                  icon="delete"
                                  onClick={() => arrayHelpers.remove(i)}
                                />
                              </Col>
                            </Row>
                          ))}
                          <Button
                            ico="plus"
                            onClick={() => {
                              arrayHelpers.push({ name: '', value: '' });
                            }}
                          >
                            {intl.get('workflow.addEnv')}
                          </Button>
                        </div>
                      )}
                    />
                  </FormItem>
                </Fragment>
              ))}
            </div>
          )}
        />
      </SectionCard>
      <SectionCard title={intl.get('output')}>
        <Field
          label={intl.get('type')}
          name="spec.pod.outputs.resources"
          required
          handleSelectChange={val => {
            setFieldValue('spec.pod.outputs.resources', [
              {
                name: '',
                type: val,
                path: '',
              },
            ]);
          }}
          component={SelectField}
        />
        <FormItem label={'Artifact'} {...defaultFormItemLayout}>
          <FieldArray
            name={'spec.pod.outputs.artifacts'}
            render={arrayHelpers => (
              <div>
                {_.get(values, 'spec.pod.outputs.artifacts', []).length > 0 && (
                  <Row gutter={16}>
                    <Col span={11}>{intl.get('name')}</Col>
                    <Col span={11}>{intl.get('path')}</Col>
                  </Row>
                )}
                {_.get(values, 'spec.pod.outputs.artifacts', []).map(
                  (a, index) => (
                    <Row key={index} gutter={16}>
                      <Col span={11}>
                        <Field
                          key={a.name}
                          name={`spec.pod.outputs.artifacts.${index}.name`}
                          component={InputField}
                          hasFeedback
                        />
                      </Col>
                      <Col span={11}>
                        <Field
                          key={a.value}
                          name={`spec.pod.outputs.artifacts.${index}.path`}
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
                  )
                )}
                <Button
                  ico="plus"
                  onClick={() => arrayHelpers.push({ name: '', path: '' })}
                >
                  {intl.get('workflow.addArtifact')}
                </Button>
              </div>
            )}
          />
        </FormItem>
      </SectionCard>
      <FormItem
        {...{
          labelCol: { span: 8 },
          wrapperCol: { span: 20 },
        }}
      >
        <Button style={{ float: 'right' }} htmlType="submit" type="primary">
          {intl.get('confirm')}
        </Button>
        <Button
          style={{ float: 'right', marginRight: 10 }}
          onClick={handleCancle}
        >
          {intl.get('cancel')}
        </Button>
      </FormItem>
    </Form>
  );
};

FormContent.propTypes = {
  history: PropTypes.object,
  values: PropTypes.object,
  handleSubmit: PropTypes.func,
  setFieldValue: PropTypes.func,
  handleCancle: PropTypes.func,
  update: PropTypes.bool,
};

export default FormContent;
