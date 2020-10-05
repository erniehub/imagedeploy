require 'yaml'
require 'json'

module Kubectl
  # This class represents auto-deploy-app chart managed by GitLab.
  class Version
    attr_reader :major, :minor, :patch

    GITLAB_MANAGED_CHART_NAME = 'auto-deploy-app'.freeze

    class << self
      # Load a chart from `helm list` output
      #
      # @param [String] data                    JSON formatted `helm list` output
      # @param [String] release_name            The release created by a chart in question
      def load_from_helm_ls(data, release_name)
        # In Helm 2, `helm ls --output json` returns an empty string when there are no releases
        return if data.empty?

        release = JSON.parse(data)['Releases'].find { |r| r['Name'] == release_name }

        return if release.nil?

        name, major, minor, patch = release['Chart'].scan(/\A(.+)-(\d+)\.(\d+)\.(\d+)/).first

        return unless gitlab_managed_chart?(name, major, minor, patch)

        self.new(major, minor, patch)
      end

      # Load a chart from Chart.yaml
      #
      # @param [String] chart_dir                The path to the chart directory
      def load_from_chart_yml(chart_dir)
        chart_hash = YAML.load_file(File.join(chart_dir, 'Chart.yaml'))

        name = chart_hash['name']
        major, minor, patch = chart_hash['version'].scan(/\A(\d+)\.(\d+)\.(\d)+/).first

        return unless gitlab_managed_chart?(name, major, minor, patch)

        self.new(major, minor, patch)
      end

      def gitlab_managed_chart?(name, major, minor, patch)
        name == GITLAB_MANAGED_CHART_NAME && major && minor && patch
      end
    end

    def initialize(major, minor, patch)
      @major = major.to_i
      @minor = minor.to_i
      @patch = patch.to_i
    end

    def to_s
      "v#{major}.#{minor}.#{patch}"
    end

    def compatible?(previous_chart)
      # v0 and v1 charts are compatible
      return true if major == 1 && previous_chart.major == 0

      major == previous_chart.major
    end

    def allowed_to_force_deploy?
      ENV["AUTO_DEVOPS_FORCE_DEPLOY_V#{major}"]
    end
  end
end
