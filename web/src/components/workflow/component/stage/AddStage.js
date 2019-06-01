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
  constructor(props) {
    super(props);
    const { templateName, values } = props;
    const currentStage = _.get(values, 'currentStage');
    const stages = _.get(values, 'stages', []);
    const modify = stages.includes(currentStage);
    this.state = {
      creationMethod: modify && !templateName ? 'custom' : 'template',
      templateData: null,
      modify,
    };
  }
  componentDidMount() {
    this.props.stageTemplate.getTemplateList();
  }

  handleChange = e => {
    const { setFieldValue, values } = this.props;
    const value = e.target.value;
    const state = { creationMethod: value };
    if (value === 'custom') {
      const currentStage = _.get(values, 'currentStage');
      const info = {
        metadata: { name: this.stageDefaultName() },
        ...customStageField,
      };
      setFieldValue(currentStage, info);
    } else {
      state.templateData = null;
    }
    this.setState(state);
  };

  transformTemplateData = (data, templateType) => {
    const value = _.cloneDeep(data);
    const _arguments = _.get(value, 'spec.pod.inputs.arguments', []);
    const commandIndex = _.findIndex(_arguments, o => o.name === 'cmd');
    if (
      commandIndex > -1 &&
      _.isString(_.get(_arguments, [commandIndex, 'value']))
    ) {
      let cmd = _arguments[commandIndex].value;
      cmd = cmd.replace(/;\s+/g, ';\n');
      _arguments[commandIndex].value = cmd;
    }

    if (templateType === 'cd') {
      const _arguments = _.get(value, 'spec.pod.inputs.arguments', []);
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
    }
    const resource = _.concat(
      _.get(value, 'spec.pod.inputs.resources', []),
      _.get(value, 'spec.pod.outputs.resources', [])
    );
    _.forEach(resource, v => {
      if (v.name) {
        v.name = '';
      }
    });
    return value;
  };

  stageDefaultName = () => {
    const { values } = this.props;
    const workflowName = _.get(values, 'metadata.name');
    const currentStageId = _.get(values, 'currentStage').split('_')[1];
    return `${workflowName}-stg${currentStageId}`;
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
    let inputs = _.merge(
      {
        metadata: {
          name: this.stageDefaultName(),
          annotations: { stageTemplate: templateType },
        },
      },
      _.pick(item, 'spec', {})
    );
    if (templateType) {
      inputs = this.transformTemplateData(inputs, templateType);
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
          const namespace = _.get(o, 'metadata.namespace', '');
          const key = namespace ? `${namespace}-${name}` : name;
          return (
            <Option value={name} key={key}>
              {_.get(o, ['metadata', 'labels', 'cyclone.dev/builtin']) ===
              'true'
                ? intl.get(`template.kinds.${name.replace('-template', '')}`) ||
                  name
                : name}
            </Option>
          );
        })}
      </Select>
    );
  };

  render() {
    const { creationMethod, templateData, modify } = this.state;
    const {
      stageTemplate: { templateList },
      values,
      project,
      update,
      setFieldValue,
    } = this.props;
    const templates = _.get(templateList, 'items');
    const currentStage = _.get(values, 'currentStage');

    if (!_.get(values, `${currentStage}`)) {
      return <Spin />;
    }
    const stageProps = {
      values,
      update,
      modify,
      project,
      setFieldValue,
    };
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
              stageId={_.get(values, 'currentStage')}
              data={templateData}
              {...stageProps}
            />
          </Fragment>
        ) : (
          <StageField {...stageProps} />
        )}
      </Form>
    );
  }
}

AddStage.propTypes = {
  stageTemplate: PropTypes.object,
  setFieldValue: PropTypes.func,
  values: PropTypes.object,
  update: PropTypes.bool,
  project: PropTypes.string,
  templateName: PropTypes.string,
};

export default AddStage;
