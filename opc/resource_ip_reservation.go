package opc

import (
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/client"
	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceOPCIPReservation() *schema.Resource {
	return &schema.Resource{
		Create: resourceOPCIPReservationCreate,
		Read:   resourceOPCIPReservationRead,
		Delete: resourceOPCIPReservationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"permanent": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
			"parent_pool": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(compute.PublicReservationPool),
				ForceNew: true,
			},
			"tags": tagsForceNewSchema(),
			"ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"used": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceOPCIPReservationCreate(d *schema.ResourceData, meta interface{}) error {

	reservation := compute.CreateIPReservationInput{
		Name:       d.Get("name").(string),
		ParentPool: compute.IPReservationPool(d.Get("parent_pool").(string)),
		Permanent:  d.Get("permanent").(bool),
	}

	tags := getStringList(d, "tags")
	if len(tags) != 0 {
		reservation.Tags = tags
	}

	computeClient, err := meta.(*Client).getComputeClient()
	if err != nil {
		return err
	}
	resClient := computeClient.IPReservations()
	info, err := resClient.CreateIPReservation(&reservation)
	if err != nil {
		return fmt.Errorf("Error creating ip reservation from parent_pool %s with tags=%s: %s",
			reservation.ParentPool, reservation.Tags, err)
	}

	d.SetId(info.Name)
	return resourceOPCIPReservationRead(d, meta)
}

func resourceOPCIPReservationRead(d *schema.ResourceData, meta interface{}) error {
	computeClient, err := meta.(*Client).getComputeClient()
	if err != nil {
		return err
	}
	resClient := computeClient.IPReservations()

	input := compute.GetIPReservationInput{
		Name: d.Id(),
	}

	result, err := resClient.GetIPReservation(&input)
	if err != nil {
		// IP Reservation does not exist
		if client.WasNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading ip reservation %s: %s", d.Id(), err)
	}

	if result == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", result.Name)
	d.Set("parent_pool", result.ParentPool)
	d.Set("permanent", result.Permanent)

	if err := setStringList(d, "tags", result.Tags); err != nil {
		return err
	}

	d.Set("ip", result.IP)
	d.Set("used", result.Used)
	return nil
}

func resourceOPCIPReservationDelete(d *schema.ResourceData, meta interface{}) error {
	computeClient, err := meta.(*Client).getComputeClient()
	if err != nil {
		return err
	}
	resClient := computeClient.IPReservations()

	input := compute.DeleteIPReservationInput{
		Name: d.Id(),
	}
	if err := resClient.DeleteIPReservation(&input); err != nil {
		return fmt.Errorf("Error deleting ip reservation %s", d.Id())
	}
	return nil
}
