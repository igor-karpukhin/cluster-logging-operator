package fluentd

import (
	"reflect"
	"testing"
	"text/template"

	"github.com/openshift/cluster-logging-operator/pkg/generators"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

func Test_pipelinesToSerializedLabels(t *testing.T) {
	type args struct {
		pipelines []logging.PipelineSpec
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "Empty pipeline spec list",
			args:    args{pipelines: []logging.PipelineSpec{}},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "Single label",
			args: args{[]logging.PipelineSpec{
				{
					OutputRefs: nil,
					InputRefs:  nil,
					Labels:     map[string]string{"field1": "label1"},
					Name:       "TestPipeline",
				},
			},
			},
			wantErr: false,
			want:    map[string]string{"TestPipeline": "{\"field1\":\"label1\"}"},
		},
		{
			name: "Multiple labels",
			args: args{[]logging.PipelineSpec{
				{
					OutputRefs: nil,
					InputRefs:  nil,
					Labels:     map[string]string{"field1": "label1", "field2": "label2"},
					Name:       "TestPipeline",
				},
			},
			},
			wantErr: false,
			want:    map[string]string{"TestPipeline": "{\"field1\":\"label1\",\"field2\":\"label2\"}"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pipelinesToSerializedLabels(tt.args.pipelines)
			if (err != nil) != tt.wantErr {
				t.Errorf("pipelinesToSerializedLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pipelinesToSerializedLabels() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigGenerator_generatePipelineNameToLabels(t *testing.T) {
	type fields struct {
		Generator                  *generators.Generator
		includeLegacyForwardConfig bool
		includeLegacySyslogConfig  bool
		useOldRemoteSyslogPlugin   bool
		storeTemplate              string
		outputTemplate             string
	}
	type args struct {
		pipelinesLabels map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Empty pipelines map",
			fields: fields{
				Generator:                  nil,
				includeLegacyForwardConfig: false,
				includeLegacySyslogConfig:  false,
				useOldRemoteSyslogPlugin:   false,
				storeTemplate:              "storeElasticsearch",
				outputTemplate:             "outputLabelConf",
			},
			args:    args{pipelinesLabels: map[string]string{}},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "Singe set of labels for single pipeline",
			fields: fields{
				Generator:                  nil,
				includeLegacyForwardConfig: false,
				includeLegacySyslogConfig:  false,
				useOldRemoteSyslogPlugin:   false,
				storeTemplate:              "storeElasticsearch",
				outputTemplate:             "outputLabelConf",
			},
			args: args{pipelinesLabels: map[string]string{"pipeline1": `{"label1": "value1"}`}},
			want: []string{`<label @pipeline1>
  <filter **>
    @type record_transformer
    <record>
      openshift { "labels": {"label1": "value1"} }
    </record>
  </filter>
</label>`},
			wantErr: false,
		},
		{
			name: "Multiple set of labels for single pipeline",
			fields: fields{
				Generator:                  nil,
				includeLegacyForwardConfig: false,
				includeLegacySyslogConfig:  false,
				useOldRemoteSyslogPlugin:   false,
				storeTemplate:              "storeElasticsearch",
				outputTemplate:             "outputLabelConf",
			},
			args: args{pipelinesLabels: map[string]string{"pipeline1": `{"label1":"value1", "label2":"value2", "label3":"value3"}`}},
			want: []string{`<label @pipeline1>
  <filter **>
    @type record_transformer
    <record>
      openshift { "labels": {"label1":"value1", "label2":"value2", "label3":"value3"} }
    </record>
  </filter>
</label>`},
			wantErr: false,
		},
		{
			name: "Multiple set of labels for multiple pipelines",
			fields: fields{
				Generator:                  nil,
				includeLegacyForwardConfig: false,
				includeLegacySyslogConfig:  false,
				useOldRemoteSyslogPlugin:   false,
				storeTemplate:              "storeElasticsearch",
				outputTemplate:             "outputLabelConf",
			},
			args: args{pipelinesLabels: map[string]string{
				"pipeline1": `{"label1":"value1", "label2":"value2", "label3":"value3"}`,
				"pipeline2": `{"label4":"value4", "label5":"value5", "label6":"value6"}`}},
			want: []string{`<label @pipeline1>
  <filter **>
    @type record_transformer
    <record>
      openshift { "labels": {"label1":"value1", "label2":"value2", "label3":"value3"} }
    </record>
  </filter>
</label>`,
				`<label @pipeline2>
  <filter **>
    @type record_transformer
    <record>
      openshift { "labels": {"label4":"value4", "label5":"value5", "label6":"value6"} }
    </record>
  </filter>
</label>`,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engn, err := generators.New("OutputLabelConf",
				&template.FuncMap{
					"labelName":           labelName,
					"sourceTypelabelName": sourceTypeLabelName,
				},
				templateRegistry...)
			if (err != nil) != tt.wantErr {
				t.Errorf("generatePipelineNameToLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			engine := &ConfigGenerator{
				Generator:                  engn,
				includeLegacyForwardConfig: tt.fields.includeLegacyForwardConfig,
				includeLegacySyslogConfig:  tt.fields.includeLegacySyslogConfig,
				useOldRemoteSyslogPlugin:   tt.fields.useOldRemoteSyslogPlugin,
				storeTemplate:              tt.fields.storeTemplate,
				outputTemplate:             tt.fields.outputTemplate,
			}

			got, err := engine.generatePipelineNameToLabels(tt.args.pipelinesLabels)
			if (err != nil) != tt.wantErr {
				t.Errorf("generatePipelineNameToLabels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generatePipelineNameToLabels() got = %v, want %v", got, tt.want)
			}
		})
	}
}
