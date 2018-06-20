package opc

import (
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/client"
	"github.com/hashicorp/go-oracle-terraform/lbaas"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceOPCLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceOPCLoadBalancerCreate,
		Read:   resourceOPCLoadBalancerRead,
		Update: resourceOPCLoadBalancerUpdate,
		Delete: resourceOPCLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // TODO name can be changed
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ip_network": {
				Type:     schema.TypeString,
				Optional: true,
				// TODO add validation for 3 part name
				// TODO add valication only supported for INTERNAL load balancer?
			},
			"premitted_methods": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				// TODO add validation
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scheme": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// TODO add validation
			},
			"server_pool": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeList, // TODO TypeSet?
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// Read only attributes
			"balancer_vips": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"canonical_host_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudgate_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceOPCLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client).lbaasClient.LoadBalancerClient()
	input := lbaas.CreateLoadBalancerInput{
		Name:   d.Get("name").(string),
		Region: d.Get("region").(string),
	}

	if description, ok := d.GetOk("description"); ok {
		input.Description = description.(string)
	}

	if enabled, ok := d.GetOk("enabled"); ok {
		input.Disabled = getDisabledStateKeyword(enabled.(bool))
	}

	if ipNetwork, ok := d.GetOk("ip_network"); ok {
		input.IPNetworkName = ipNetwork.(string)
	}

	if scheme, ok := d.GetOk("scheme"); ok {
		input.Scheme = lbaas.LoadBalancerScheme(scheme.(string))
	}

	if serverPool, ok := d.GetOk("server_pool"); ok {
		input.ServerPool = serverPool.(string)
	}

	// TODO permittedMethods := getStringList(d, "premitted_methods")
	// if len(permittedMethods) != 0 {
	// 	input.PermittedMethods = permittedMethods
	// }

	tags := getStringList(d, "tags")
	if len(tags) != 0 {
		input.Tags = tags
	}

	info, err := client.CreateLoadBalancer(&input)
	if err != nil {
		return fmt.Errorf("Error creating Load Balancer: %s", err)
	}

	d.SetId(info.Name)
	return resourceOPCLoadBalancerRead(d, meta)
}

func resourceOPCLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	lbaasClient := meta.(*Client).lbaasClient.LoadBalancerClient()
	name := d.Id()
	region := d.Get("region").(string)

	result, err := lbaasClient.GetLoadBalancer(region, name)
	if err != nil {
		// LoadBalancer does not exist
		if client.WasNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Load Balancer %s: %s", d.Id(), err)
	}

	if result == nil {
		d.SetId("")
		return nil
	}

	d.Set("canonical_host_name", result.CanonicalHostName)
	d.Set("cloudgate_capable", result.CloudgateCapable)
	d.Set("enabled", getEnabledState(result.Disabled))
	d.Set("description", result.Description)
	d.Set("ip_network", result.IPNetworkName)
	d.Set("name", result.Name)
	d.Set("region", result.Region)
	d.Set("scheme", result.Scheme)
	d.Set("server_pool", result.ServerPool)
	d.Set("uri", result.URI)

	if err := setStringList(d, "balancer_vips", result.BalancerVIPs); err != nil {
		return err
	}
	if err := setStringList(d, "premitted_methods", result.PermittedMethods); err != nil {
		return err
	}
	if err := setStringList(d, "tags", result.Tags); err != nil {
		return err
	}
	return nil
}

func resourceOPCLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	lbaasClient := meta.(*Client).lbaasClient.LoadBalancerClient()
	name := d.Id()
	region := d.Get("region").(string)

	input := lbaas.UpdateLoadBalancerInput{}

	if description, ok := d.GetOk("description"); ok {
		input.Description = description.(string)
	}

	if enabled, ok := d.GetOk("enabled"); ok {
		input.Disabled = getDisabledStateKeyword(enabled.(bool))
	}

	if ipNetwork, ok := d.GetOk("ip_network"); ok {
		input.IPNetworkName = ipNetwork.(string)
	}

	if serverPool, ok := d.GetOk("server_pool"); ok {
		input.ServerPool = serverPool.(string)
	}

	permittedMethods := getStringList(d, "premitted_methods")
	if len(permittedMethods) != 0 {
		input.PermittedMethods = permittedMethods
	}

	tags := getStringList(d, "tags")
	if len(tags) != 0 {
		input.Tags = tags
	}

	if description, ok := d.GetOk("description"); ok {
		input.Description = description.(string)
	}

	result, err := lbaasClient.UpdateLoadBalancer(region, name, &input)
	if err != nil {
		return fmt.Errorf("Error updating LoadBalancer: %s", err)
	}

	d.SetId(result.Name)

	// TODO instead of re-read, process info from UpdateLoadBalancer()
	return resourceOPCLoadBalancerRead(d, meta)
}

func resourceOPCLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	lbaasClient := meta.(*Client).lbaasClient.LoadBalancerClient()
	name := d.Id()
	region := d.Get("region").(string)

	if _, err := lbaasClient.DeleteLoadBalancer(region, name); err != nil {
		return fmt.Errorf("Error deleting LoadBalancer")
	}
	return nil
}

// return the Disbaled State keyword for Load Balancer enabled state
func getDisabledStateKeyword(enabled bool) lbaas.LoadBalancerDisabled {
	if enabled {
		return lbaas.LoadBalancerDisabledFalse
	} else {
		return lbaas.LoadBalancerDisabledTrue
	}
}

// convert the DisabledState attribute to a boolean representing the enabled state
func getEnabledState(state lbaas.LoadBalancerDisabled) bool {
	if state == lbaas.LoadBalancerDisabledFalse {
		return false
	}
	return true
}
