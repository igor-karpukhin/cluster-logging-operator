package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"

	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Generating fluentd config", func() {
	var (
		outputs       []loggingv1.OutputSpec
		forwarderSpec *loggingv1.ForwarderSpec
		generator     *ConfigGenerator
	)

	Context("for cloudwatch output ", func() {
		BeforeEach(func() {
			var err error
			generator, err = NewConfigGenerator(false, false, true)
			Expect(err).To(BeNil())
			Expect(generator).ToNot(BeNil())

			outputs = []loggingv1.OutputSpec{
				{
					Type: loggingv1.OutputTypeCloudwatch,
					Name: "my-cloudwatch",
					OutputTypeSpec: loggingv1.OutputTypeSpec{
						Cloudwatch: &loggingv1.Cloudwatch{
							Region: "anumber1",
							LogStreamStrategy: loggingv1.CloudwatchLogStreamStrategy{
								Name: loggingv1.LogStreamStrategyTypeUnique,
								CloudwatchLogStreamStrategyTypeSpec: loggingv1.CloudwatchLogStreamStrategyTypeSpec{
									RetentionInDays: 7,
								},
							},
						},
					},
					Secret: &loggingv1.OutputSecretSpec{
						Name: "my-secret",
					},
				},
			}
		})

		It("should provide a valid configuration", func() {
			expConf := `
			<label @MY_CLOUDWATCH>
				<filter **>
					@type record_transformer
					<record>
						cw_stream_name ${tag}
						cw_retention_days 7
					</record>
				</filter>
				<match **>
					@type cloudwatch_logs
					auto_create_stream true
					region anumber1
					log_group_name openshiftlogging
					log_stream_name_key cw_stream_name
					remove_log_stream_name_key true
					auto_create_stream true
					concurrency 2
					aws_key_id /var/run/ocp-collector/secrets/my-secret/aws_access_key_id
					aws_sec_key /var/run/ocp-collector/secrets/my-secret/aws_secret_access_key
					retention_in_days_key cw_retention_days
					#max_message_length 32768
					#use_tag_as_group false
					#use_tag_as_stream false
					include_time_key true
					#localtime true
					#log_group_name_key group_name_key
					#remove_log_group_name_key true
					#put_log_events_retry_wait 1s
					#put_log_events_retry_limit 17
					#put_log_events_disable_retry_limit false
					log_rejected_request true
				</match>
			</label>`

			results, err := generator.generateOutputLabelBlocks(outputs, forwarderSpec)
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(1))
			Expect(results[0]).To(EqualTrimLines(expConf))
		})
	})
})
