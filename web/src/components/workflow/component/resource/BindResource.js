import { Form, Input, Radio } from 'antd';
import { Field, FieldArray } from 'formik';
import MakeField from '@/components/public/makeField';
import { noLabelItemLayout, modalFormItemLayout } from '@/lib/const';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';
import SelectPlus from '@/components/public/makeField/select';
import SCM from './SCM';

const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;
const FormItem = Form.Item;
const InputField = MakeField(Input);
const SelectField = MakeField(SelectPlus);
const Fragment = React.Fragment;

const inputArray = ['SCM', 'DockerRegistry', 'Cluster', 'SonarQube'];

// use in add stage, select a exist resource or create a new resource
@inject('integration')
@observer
class BindResource extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    values: PropTypes.object,
    type: PropTypes.oneOf(['inputs', 'outputs']),
    integration: PropTypes.object,
  };

  state = { addWay: 'new' };

  componentDidMount() {
    const { integration } = this.props;
    integration.getIntegrationList();
  }

  changeAddWay = value => {
    this.setState({ addWay: value });
  };

  render() {
    const { addWay } = this.state;
    const { values, setFieldValue, type, integration } = this.props;
    const resourceType = _.get(values, 'resourceType', 'SCM');
    const resourceList = _.get(
      integration,
      `groupIntegrationList.${resourceType}`
    );
    // TODO(qme): i18n
    return (
      <Form layout={'horizontal'}>
        {type === 'inputs' ? (
          <Field
            label={intl.get('type')}
            name="resourceType"
            required
            handleSelectChange={val => {
              setFieldValue('resourceType', val);
              // TODO(qme): set value
              // setFieldValue('spec.parameters', resourceParametersField[val]);
            }}
            payload={{
              items: inputArray,
            }}
            component={SelectField}
            formItemLayout={modalFormItemLayout}
          />
        ) : (
          <FormItem
            label={intl.get('workflow.resourceType')}
            {...modalFormItemLayout}
          >
            image
          </FormItem>
        )}
        {type === 'inputs' && (
          <FormItem
            label={intl.get('workflow.addMethod')}
            {...modalFormItemLayout}
          >
            <RadioGroup onChange={this.changeAddWay} defaultValue={addWay}>
              <RadioButton value="new">{intl.get('operation.add')}</RadioButton>
              <RadioButton value="exist">
                {intl.get('workflow.existResource')}
              </RadioButton>
            </RadioGroup>
          </FormItem>
        )}
        {addWay === 'exist' ? (
          <Field
            label={intl.get('type')}
            name="matadata.name"
            required
            handleSelectChange={val => {
              setFieldValue('name', val);
            }}
            component={<div>TODO: resource select</div>}
            formItemLayout={modalFormItemLayout}
          />
        ) : (
          <Fragment>
            <Field
              label={intl.get('name')}
              name="matadata.name"
              component={InputField}
              formItemLayout={modalFormItemLayout}
              hasFeedback
              required
            />
            {resourceType === 'SCM' ? (
              <SCM
                values={values}
                integrationList={resourceList}
                setFieldValue={setFieldValue}
              />
            ) : (
              <FormItem {...noLabelItemLayout}>
                <FieldArray
                  name="spec.parameters"
                  render={() => (
                    <div>
                      {_.get(values, 'spec.parameters', []).map(
                        (field, index) => (
                          <Field
                            key={field.name}
                            label={
                              field.name.includes('GIT_')
                                ? field.name.replace('GIT_', '')
                                : field.name
                            }
                            name={`spec.parameters.${index}.value`}
                            component={InputField}
                            formItemLayout={modalFormItemLayout}
                            hasFeedback
                            required
                          />
                        )
                      )}
                    </div>
                  )}
                />
              </FormItem>
            )}
          </Fragment>
        )}
        {type === 'inputs' && (
          <Field
            label={intl.get('workflow.usePath')}
            name="path"
            component={InputField}
            hasFeedback
            required
            formItemLayout={modalFormItemLayout}
          />
        )}
      </Form>
    );
  }
}

export default BindResource;
