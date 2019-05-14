import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import { Form, Radio, Select, Spin } from 'antd';
import { defaultFormItemLayout, customStageField } from '@/lib/const';
import StageField from './StageField';
import TemplateStage from './TemplateStage';

const Option = Select.Option;
const FormItem = Form.Item;
const Fragment = React.Fragment;

@inject('stageTemplate')
@observer
class AddStage extends React.Component {
  componentDidMount() {
    this.props.stageTemplate.getTemplateList();
  }
  state = {
    creationMethod: 'custom',
    templateData: null,
  };

  handleChange = e => {
    const { setFieldValue, values } = this.props;
    const value = e.target.value;
    if (value === 'custom') {
      const currentStage = _.get(values, 'currentStage');
      setFieldValue(currentStage, customStageField);
    }
    this.setState({ creationMethod: value });
  };

  transformTemplateData = data => {
    const value = _.cloneDeep(data);
    const _arguments = _.get(data, 'spec.pod.inputs.arguments', []);

    _.forEach(_arguments, (v, k) => {
      if (v.name === 'config') {
        v.value = {
          namespace: '',
          name: '',
          container: '',
          image: '',
        };
      }
    });
    value.spec.pod.inputs.arguments = _arguments;
    return value;
  };

  // Select template then render field
  selectTemplate = value => {
    const {
      stageTemplate: { templateList },
      setFieldValue,
      values,
    } = this.props;
    const item = _.find(
      _.get(templateList, 'items', []),
      o => _.get(o, 'metadata.name') === value
    );
    const templateType = _.get(item, [
      'metadata',
      'labels',
      'stage.cyclone.dev/template-kind',
    ]);
    let inputs = _.merge({ name: '' }, _.pick(item, 'spec', {}));
    if (templateType === 'cd') {
      inputs = this.transformTemplateData(inputs);
    }
    this.setState({ templateData: inputs });
    const currentStage = _.get(values, 'currentStage');

    setFieldValue(currentStage, inputs);
  };

  renderTemplateSelect = templates => {
    const { templateData } = this.state;
    if (!templates) {
      return <Spin />;
    }

    const defaultValue = _.get(templates, '[0].metadata.name');
    if (!templateData) {
      this.selectTemplate(defaultValue);
    }
    return (
      <Select
        showSearch
        onSelect={this.selectTemplate}
        defaultValue={defaultValue}
      >
        {templates.map(o => {
          const name = _.get(o, 'metadata.name');
          return (
            <Option value={name} key={name}>
              {_.get(o, ['metadata', 'labels', 'cyclone.dev/builtin']) ===
              'true'
                ? intl.get(`template.kinds.${name.replace('-template', '')}`)
                : _.get(o, ['metadata', 'annotations', 'cyclone.dev/alias'])}
            </Option>
          );
        })}
      </Select>
    );
  };

  render() {
    const { creationMethod, templateData } = this.state;
    const {
      stageTemplate: { templateList },
      values,
    } = this.props;
    const templates = _.get(templateList, 'items');
    const currentStage = _.get(values, 'currentStage');
    const stages = _.get(values, 'stages', []);
    const modify = stages.includes(currentStage);
    return (
      <Form>
        {!modify && (
          <FormItem
            label={intl.get('workflow.stageCreation')}
            {...defaultFormItemLayout}
          >
            <Radio.Group
              onChange={this.handleChange}
              defaultValue={creationMethod}
              value={creationMethod}
            >
              <Radio.Button value="template">
                {intl.get('sideNav.stageTemplate')}
              </Radio.Button>
              <Radio.Button value="custom">{intl.get('custom')}</Radio.Button>
            </Radio.Group>
          </FormItem>
        )}
        {creationMethod === 'template' ? (
          <Fragment>
            {!modify && (
              <FormItem
                label={intl.get('workflow.selectTemplate')}
                {...defaultFormItemLayout}
              >
                {this.renderTemplateSelect(templates)}
              </FormItem>
            )}
            <TemplateStage
              values={values}
              stageId={_.get(values, 'currentStage')}
              data={templateData}
            />
          </Fragment>
        ) : (
          <StageField values={this.props.values} />
        )}
      </Form>
    );
  }
}

AddStage.propTypes = {
  stageTemplate: PropTypes.object,
  setFieldValue: PropTypes.func,
  values: PropTypes.object,
};

export default AddStage;
