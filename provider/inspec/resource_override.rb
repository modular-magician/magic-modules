# Copyright 2017 Google Inc.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

require 'provider/core'
require 'provider/overrides/resources'

module Provider
  class Inspec < Provider::Core
    # Custom inspec code to handle type convergence operations
    class Handlers < Api::Object
      def validate
        super
      end
    end

    # Product specific overriden properties for inspec
    class ResourceOverride < Provider::Overrides::ResourceOverride
      def self.attributes
        [:manual]
      end

      attr_reader(*attributes)

      def validate
        assign_defaults

        super
        check_property :manual, :boolean
      end

      private

      def assign_defaults
        default_value_property :manual, false
      end
    end
  end
end
