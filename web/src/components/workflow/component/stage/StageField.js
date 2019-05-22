import { Input, Form, Button, Spin, Row, Col } from 'antd';
import MakeField from '@/components/public/makeField';
import { Field, FieldArray } from 'formik';
import SectionCard from '@/components/public/sectionCard';
import { defaultFormItemLayout } from '@/lib/const';
import ResourceArray from '../resource/ResourceArray';
import PropTypes from 'prop-types';

const Fragment = React.Fragment;
const FormItem = Form.Item;
const { TextArea } = Input;

const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

class StageField extends React.Component {
  static propTypes = {
    values: PropTypes.object,
    update: PropTypes.boolean,
    project: PropTypes.string,
    modify: PropTypes.boolean,
  };

  state = {
    visible: false,
  };

  render() {
    const { values, update, project, modify } = this.props;
    const currentStage = _.get(values, 'currentStage');
    if (!currentStage) {
      return <Spin />;
    }
    const specKey = `${currentStage}.spec.pod`;
    return (
      <Fragment>
        {update && modify ? (
          <FormItem label={intl.get('name')} {...defaultFormItemLayout}>
            {_.get(values, `${currentStage}.metadata.name`)}
          </FormItem>
        ) : (
          <Field
            label={intl.get('name')}
            name={`${currentStage}.metadata.name`}
            component={InputField}
            hasFeedback
            required
          />
        )}
        <SectionCard title={intl.get('input')}>
          <ResourceArray
            resourcesField={`${specKey}.inputs.resources`}
            resources={_.get(values, `${specKey}.inputs.resources`, [])}
            update={update}
            project={project}
          />
        </SectionCard>
        <SectionCard title={intl.get('config')}>
          <FieldArray
            name={`${specKey}.spec.containers`}
            render={() => (
              <div>
                {_.get(values, `${specKey}.spec.containers`, []).map(
                  (a, index) => (
                    <Fragment key={index}>
                      <Field
                        label={intl.get('image')}
                        name={`${specKey}.spec.containers.${index}.image`}
                        component={InputField}
                        hasFeedback
                        required
                      />
                      {/* // TODO: 暂时不展示此字段 */}
                      {/* <Field
                        label={'ENTRYPOINT'}
                        name={`${specKey}.spec.containers.${index}.command`}
                        component={TextareaField}
                        hasFeedback
                        required
                      /> */}
                      <Field
                        label={'COMMAND'}
                        name={`${specKey}.spec.containers.${index}.args`}
                        component={TextareaField}
                        hasFeedback
                        required
                      />
                      <FormItem
                        label={intl.get('env')}
                        {...defaultFormItemLayout}
                      >
                        <FieldArray
                          name={`${specKey}.spec.containers.${index}.env`}
                          render={arrayHelpers => (
                            <div>
                              {_.get(
                                values,
                                `${specKey}.spec.containers.${index}.env`,
                                []
                              ).length > 0 && (
                                <Row gutter={16}>
                                  <Col span={11}>{intl.get('key')}</Col>
                                  <Col span={11}>{intl.get('value')}</Col>
                                </Row>
                              )}
                              {_.get(
                                values,
                                `${specKey}.spec.containers.${index}.env`,
                                []
                              ).map((a, i) => (
                                <Row key={i} gutter={16}>
                                  <Col span={11}>
                                    <Field
                                      key={`env_name_${i}`}
                                      name={`${specKey}.spec.containers.${index}.env.${i}.name`}
                                      component={InputField}
                                      hasFeedback
                                    />
                                  </Col>
                                  <Col span={11}>
                                    <Field
                                      key={`env_value_${i}`}
                                      name={`${specKey}.spec.containers.${index}.env.${i}.value`}
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
                                onClick={() =>
                                  arrayHelpers.push({ name: '', value: '' })
                                }
                              >
                                {intl.get('workflow.addEnv')}
                              </Button>
                            </div>
                          )}
                        />
                      </FormItem>
                    </Fragment>
                  )
                )}
              </div>
            )}
          />
        </SectionCard>
        <SectionCard title={intl.get('output')}>
          <ResourceArray
            type="outputs"
            resourcesField={`${currentStage}.outputs.resources`}
            resources={_.get(values, `${currentStage}.outputs.resources`, [])}
          />
          {/* // NOTE: temporarily not supported artifacts */}
          {/* <FormItem label={'artifacts'} {...defaultFormItemLayout}>
            <FieldArray
              name={`${currentStage}.outputs.artifacts`}
              render={arrayHelpers => (
                <div>
                  {_.get(values, `${currentStage}.outputs.artifacts`, [])
                    .length > 0 && (
                    <Row gutter={16}>
                      <Col span={11}>{intl.get('name')}</Col>
                      <Col span={11}>{intl.get('path')}</Col>
                    </Row>
                  )}
                  {_.get(values, `${currentStage}.outputs.artifacts`, []).map(
                    (r, i) => (
                      <Row key={i} gutter={16}>
                        <Col span={11}>
                          <Field
                            key={`artifacts_name_${i}`}
                            name={`${currentStage}.outputs.artifacts.${i}.name`}
                            component={InputField}
                            hasFeedback
                          />
                        </Col>
                        <Col span={11}>
                          <Field
                            key={`artifacts_path_${i}`}
                            name={`${currentStage}.outputs.artifacts.${i}.path`}
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
          </FormItem> */}
        </SectionCard>
      </Fragment>
    );
  }
}

export default StageField;
