// functions to plot econdata series

package plot

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fabiansalazares/bdsicego/series"

	//"github.com/gonum/plot"
	//"gonum.org/v1/plot"
	//"gonum.org/v1/plotter"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	//	"gonum.org/v1/plotter"
)

func getXYPoints(serie series.EconSerie) *plotter.XYs {
	data := serie.GetData()

	points := make(plotter.XYs, len(data.Dates))

	for i := range points {
		points[i].X = float64(data.Dates[i].Unix())
		points[i].Y = data.Values[i]
	}

	return &points
}

func Plot(seriesToPlot ...series.BDSICESerie) (string, error) {
	// array of pointers to XYs objects that contain the set of X and Y points for each observation

	var (
		seriesPlotters []interface{}
		seriesPoints   []*plotter.XYs
	)

	// create p object -> the plot
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	// possible addgrid (should be set up either as a function argument, or as a config option)
	p.Add(plotter.NewGrid())

	if len(seriesToPlot) > 1 {
		// if there is only serie to be plotted, the y-axis can be labeled safely to the unit of the serie to plot
		p.Title.Text = fmt.Sprintf("%s - %s", seriesToPlot[0].GetCode(), seriesToPlot[0].GetTitle())
		p.Y.Label.Text = seriesToPlot[0].GetUnit()

	} else {

		// if there is more than one serie to plot, we need to check if all the series have the same units
		// if they do have the same units, the y-axis will be labelled accordingly. Otherwise, an empty label will be added
		differentUnits := false

		for i := 1; i < len(seriesToPlot); i++ {
			if seriesToPlot[i].GetUnit() != seriesToPlot[0].GetUnit() {
				differentUnits = true
			}
		}

		if differentUnits {
			p.Title.Text = ""
		} else {
			p.Title.Text = seriesToPlot[0].GetTitle()
			p.Y.Label.Text = seriesToPlot[0].GetUnit()
		}
	}

	// x-axis label will always be time since we are dealing with time series
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-1"}
	//p.X.Label.Text = "t"

	// create an array of XYs objects for each serie to plot
	for i := 0; i < len(seriesToPlot); i++ {
		seriesPoints = append(seriesPoints, getXYPoints(seriesToPlot[i]))
	}

	// add plotters to the plot p

	// plotutil.AddLinePoints nicely cycles through color codes assigning one to each line plotter object
	// that we add to the plot object. It takes as arguments the plot to which it will be assigned,
	// and then alternatively the LinePoints title and the XYs of the LinePoints. This is a very poor
	// design choice because it is fully dependant on the order in which we pass the arguments
	// So, in order to make it work, we need to create a slice of empty interface objects
	// to which we will add alternatively the string containing the series name and the XYs themselves.
	// For now, it works, but a better solution should be found in the future, also considering
	// posible options to set the plotting style.

	for i := 0; i < len(seriesToPlot); i++ {

		seriesPlotters = append(seriesPlotters, seriesToPlot[i].GetTitle())
		seriesPlotters = append(seriesPlotters, seriesPoints[i])
	}

	err = plotutil.AddLinePoints(p, seriesPlotters...)
	if err != nil {
		return "", fmt.Errorf("econdata/plot: an error ocurred adding line and points to the plot: %s", err.Error())
	}

	// create a tmp file to save the plot to. Actually, we don't want the file handler but just the file name

	tmp, err := ioutil.TempFile(os.TempDir(), "econdata-plot*.png")
	if err != nil {
		return "", fmt.Errorf("econdata/plot: could not generate tmp file: %s", err.Error())
	}

	// save the plot to the temporal file that we just created.
	if err := p.Save(10*vg.Inch, 4*vg.Inch, tmp.Name()); err != nil {
		return "", fmt.Errorf("econdata/plot: an error ocurred saving the plot to tmp file %s: %s", tmp.Name(), err.Error())
	}

	// we return the name of the plot. plotCommand should be plotting
	return tmp.Name(), nil

}
