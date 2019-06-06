import { Input, Form } from 'antd';
import { Field, FieldArray } from 'formik';
import SectionCard from '@/components/public/sectionCard';
import MakeField from '@/components/public/makeField';
import ResourceArray from '../resource/ResourceArray';
import { required } from '@/components/public/validate';
import { drawerFormItemLayout } from '@/lib/const';
import style from '@/components/workflow/component/index.module.less';
import PropTypes from 'prop-types';

const Fragment = React.Fragment;
const { TextArea } = Input;
const FormItem = Form.Item;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

class TemplateStage extends React.Component {
  static propTypes = {
    stageId: PropTypes.string,
    values: PropTypes.object,
    update: PropTypes.bool,
    project: PropTypes.string,
    modify: PropTypes.bool,
    argDes: PropTypes.object,
  };

  renderSection = (data, key) => {
    const dom = [];
    _.forEach(data.value, (v, k) => {
      dom.push(
        <Field
          key={key}
          label={k}
          name={`${key}.${k}`}
          component={InputField}
          hasFeedback
          required
          validate={required}
          formItemLayout={drawerFormItemLayout}
        />
      );
    });
    return dom;
  };
  render() {
    const { stageId, values, update, project, modify, argDes } = this.props;
    const specKey = `${stageId}.spec.pod`;
    const outputResource = _.get(values, `${specKey}.outputs.resources`);
    const resourceProps = {
      update,
      projectName: project,
    };
    return (
      <Fragment>
        {update && modify ? (
          <FormItem label={intl.get('name')} {...drawerFormItemLayout}>
            {_.get(values, `${stageId}.metadata.name`)}
          </FormItem>
        ) : (
          <Field
            label={intl.get('name')}
            name={`${stageId}.metadata.name`}
            component={InputField}
            hasFeedback
            required
            validate={required}
            formItemLayout={drawerFormItemLayout}
          />
        )}
        <SectionCard title={intl.get('input')}>
          <ResourceArray
            resourcesField={`${specKey}.inputs.resources`}
            resources={_.get(values, `${specKey}.inputs.resources`, [])}
            {...resourceProps}
          />
          <div className={style['divider-small']}>Arguments</div>
          <FieldArray
            name={`${specKey}.inputs.arguments`}
            render={arrayHelpers => (
              <div>
                {_.get(values, `${specKey}.inputs.arguments`, []).map(
                  (r, i) => {
                    if (_.isObject(r.value)) {
                      return this.renderSection(
                        r,
                        `${specKey}.inputs.arguments.${i}.value`
                      );
                    } else {
                      return (
                        <Field
                          key={i}
                          label={r.name}
                          name={`${specKey}.inputs.arguments.${i}.value`}
                          component={
                            ['cmd'].includes(r.name)
                              ? TextareaField
                              : InputField
                          }
                          tooltip={_.get(argDes, r.name)}
                          hasFeedback
                          required
                          validate={required}
                          formItemLayout={drawerFormItemLayout}
                        />
                      );
                    }
                  }
                )}
              </div>
            )}
          />
        </SectionCard>
        {outputResource && (
          <SectionCard title={intl.get('output')}>
            <ResourceArray
              resourcesField={`${specKey}.outputs.resources`}
              resources={_.get(values, `${specKey}.outputs.resources`, [])}
              type="outputs"
              {...resourceProps}
            />
          </SectionCard>
        )}
      </Fragment>
    );
  }
}

export default TemplateStage;
