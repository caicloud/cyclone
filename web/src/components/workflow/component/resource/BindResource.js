import { Form, Input, Radio, Select } from 'antd';
import { Field, FieldArray } from 'formik';
import MakeField from '@/components/public/makeField';
import {
  defaultFormItemLayout,
  noLabelItemLayout,
  resourceParametersField,
} from '@/lib/const';
import PropTypes from 'prop-types';

const Option = Select.Option;
const RadioButton = Radio.Button;
const RadioGroup = Radio.Group;
const FormItem = Form.Item;
const InputField = MakeField(Input);
const Fragment = React.Fragment;

const inputArray = ['scm', 'image', 'docker registry'];

// use in add stage, select a exist resource or create a new resource
class BindResource extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    values: PropTypes.object,
    type: PropTypes.oneOf(['input', 'output']),
  };

  state = { addWay: 'new' };

  changeResourceType = value => {
    const { setFieldValue } = this.props;
    setFieldValue('spec.parameters', resourceParametersField[value]);
  };

  changeAddWay = value => {
    this.setState({ addWay: value });
  };

  render() {
    const { addWay } = this.state;
    const { values, setFieldValue, type } = this.props;
    // TODO(qme): i18n
    return (
      <Form layout={'horizontal'}>
        <FormItem
          label={intl.get('workflow.resourceType')}
          {...defaultFormItemLayout}
        >
          {type === 'input' ? (
            <Select onSelect={this.changeResourceType} defaultValue="scm">
              {inputArray.map(o => (
                <Option value={o} key={o}>
                  {o}
                </Option>
              ))}
            </Select>
          ) : (
            'image'
          )}
        </FormItem>
        {type === 'input' && (
          <FormItem
            label={intl.get('workflow.addMethod')}
            {...defaultFormItemLayout}
          >
            {/* TODO:换成 select */}
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
            name="name"
            required
            handleSelectChange={val => {
              setFieldValue('name', val);
            }}
            component={<div>TODO: resource select</div>}
          />
        ) : (
          <Fragment>
            <Field
              label={intl.get('name')}
              name="name"
              component={InputField}
              hasFeedback
              required
            />
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
                          hasFeedback
                          required
                        />
                      )
                    )}
                  </div>
                )}
              />
            </FormItem>
          </Fragment>
        )}
        {type === 'input' && (
          <Field
            label={intl.get('workflow.usePath')}
            name="path"
            component={InputField}
            hasFeedback
            required
          />
        )}
      </Form>
    );
  }
}

export default BindResource;
